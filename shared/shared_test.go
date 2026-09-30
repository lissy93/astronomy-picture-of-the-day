package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// stubNASA serves a settable feed body plus /img and /thumb, and counts feed hits.
func stubNASA(t *testing.T) (base string, hits *int32, setBody func(string)) {
	t.Helper()
	var count int32
	var body atomic.Value
	body.Store("[]")

	mux := http.NewServeMux()
	mux.HandleFunc("/apod", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		if r.URL.Query().Get("per_page") != "1" {
			t.Errorf("feed requested without per_page=1: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body.Load().(string))
	})
	mux.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		io.WriteString(w, "IMAGE-BYTES")
	})
	mux.HandleFunc("/thumb", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		io.WriteString(w, "THUMB-BYTES")
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL, &count, func(s string) { body.Store(s) }
}

func newService(base string) *Service {
	return New(&Config{
		NASABaseURL: base + "/apod",
		CacheTTL:    time.Minute,
	})
}

// feed wraps a single entry in the array the apod-basic feed returns.
func feed(entry map[string]string) string {
	b, _ := json.Marshal([]map[string]string{entry})
	return string(b)
}

func TestFetchMapsImageDay(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(feed(map[string]string{
		"date":        "2026-09-30",
		"title":       "Arp 78: Peculiar Galaxy in Aries",
		"media_type":  "image",
		"explanation": `<strong>Explanation:</strong> <a href="x">Peculiar</a> spiral galaxy <a href="y">Arp 78</a> .<br><br><strong>APOD's email has changed.</strong> See <a href="z">here</a>.<br><strong>Tomorrow's picture: </strong>a harvest`,
		"copyright":   `<a href="a">Karl Glazebrook</a> &amp; <a href="b">Ivan Baldry</a> (<a href="c">JHU</a>)`,
		"url":         "https://science.nasa.gov/image-article/apod-2026-september-30/",
		"hdurl":       "https://assets.science.nasa.gov/dynamicimage/assets/apod/NGC772.jpg?w=1772&h=1182&fit=clip",
	}))

	got, err := newService(base).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	want := Response{
		Date:        "2026-09-30",
		Title:       "Arp 78: Peculiar Galaxy in Aries",
		MediaType:   "image",
		Explanation: "Peculiar spiral galaxy Arp 78.",
		Copyright:   "Karl Glazebrook & Ivan Baldry (JHU)",
		URL:         "https://assets.science.nasa.gov/dynamicimage/assets/apod/NGC772.jpg",
		HdURL:       "https://assets.science.nasa.gov/dynamicimage/assets/apod/NGC772.jpg?w=1772&h=1182&fit=clip",
	}
	if *got != want {
		t.Errorf("got  %+v\nwant %+v", *got, want)
	}
}

func TestFetchMapsVideoDays(t *testing.T) {
	cases := map[string]struct{ html, want string }{
		"mp4": {
			`<video controls title=""><source src="https://assets.science.nasa.gov/content/dam/apod/xz_and.mp4" type="video/mp4"></video>`,
			"https://assets.science.nasa.gov/content/dam/apod/xz_and.mp4",
		},
		"youtube": {
			`<iframe width="960" height="540" src="//www.youtube.com/embed/UgxWkOXcdZU?rel=0&amp;t=1" frameborder="0" allowfullscreen></iframe>`,
			"https://www.youtube.com/embed/UgxWkOXcdZU?rel=0&t=1",
		},
		"lazy-loaded": {`<video data-src="lazy.mp4"><source src='real.mp4'></video>`, "real.mp4"},
		"none":        {`<p>No player here</p>`, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			base, _, setBody := stubNASA(t)
			setBody(feed(map[string]string{
				"media_type": "video",
				"hdurl":      "https://assets.science.nasa.gov/dynamicimage/assets/apod/frame.jpg?w=1920",
				"basic_html": tc.html,
			}))

			got, err := newService(base).Fetch(context.Background())
			if err != nil {
				t.Fatalf("Fetch: %v", err)
			}
			if got.URL != tc.want {
				t.Errorf("url = %q, want %q", got.URL, tc.want)
			}
			if got.ThumbnailURL != "https://assets.science.nasa.gov/dynamicimage/assets/apod/frame.jpg" {
				t.Errorf("thumbnail_url = %q", got.ThumbnailURL)
			}
			if got.HdURL != "" {
				t.Errorf("video days should omit hdurl, got %q", got.HdURL)
			}
		})
	}
}

func TestCleanExplanation(t *testing.T) {
	cases := []struct{ in, want string }{{
		`<strong>Explanation:&nbsp;</strong>Rock or ice.<br><br>Look up!<br><br><strong>Tomorrow's picture:</strong> a red Sun`,
		"Rock or ice. Look up!", // an unlabelled paragraph is kept
	}, {
		`<strong>Explanation: </strong>Toward Perseus.<br><strong><br>Gallery: </strong><a href="x">2026</a>`,
		"Toward Perseus.",
	}, {
		`<b>Explanation:</b> Ice.<br><strong class="note">APOD has moved</strong>`,
		"Ice.",
	}, {
		"<strong>Explanation:\u00a0</strong>No footer at all ( <a>Pisces</a> ).",
		"No footer at all (Pisces).",
	}}
	for _, tc := range cases {
		if got := cleanExplanation(tc.in); got != tc.want {
			t.Errorf("cleanExplanation(%q)\n got %q\nwant %q", tc.in, got, tc.want)
		}
	}
}

func TestCleanCredit(t *testing.T) {
	cases := []struct{ in, want string }{
		{`<b> Image Credit: </b> <a href="x">Apollo 11</a> , NASA<br><em>Processing:</em> J. Miller`, "Apollo 11, NASA Processing: J. Miller"},
		{`Video Credit &amp; Copyright: Cassini Imaging Team, ISS`, "Cassini Imaging Team, ISS"},
		{`<a href="x">Robert Eder</a>`, "Robert Eder"},
	}
	for _, tc := range cases {
		if got := cleanCredit(tc.in); got != tc.want {
			t.Errorf("cleanCredit(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFetchEmptyFeed(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody("[]")
	if _, err := newService(base).Fetch(context.Background()); err == nil {
		t.Fatal("expected an error for an empty feed")
	}
}

func TestFetchCaches(t *testing.T) {
	base, hits, setBody := stubNASA(t)
	setBody(feed(map[string]string{"media_type": "image", "title": "Cached"}))

	svc := newService(base)
	for i := 0; i < 3; i++ {
		if _, err := svc.Fetch(context.Background()); err != nil {
			t.Fatalf("Fetch %d: %v", i, err)
		}
	}
	if n := atomic.LoadInt32(hits); n != 1 {
		t.Errorf("upstream hit %d times, want 1 (cache should absorb the rest)", n)
	}
}

func TestHandleApod(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(feed(map[string]string{"media_type": "image", "title": "JSON Day", "hdurl": "u"}))

	rr := httptest.NewRecorder()
	newService(base).HandleApod()(rr, httptest.NewRequest(http.MethodGet, "/apod", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("missing Cache-Control, got %q", cc)
	}
	var got Response
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Title != "JSON Day" {
		t.Errorf("title = %q", got.Title)
	}
}

func TestHandleImageServesImage(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(feed(map[string]string{"media_type": "image", "hdurl": base + "/img"}))

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if body := rr.Body.String(); body != "IMAGE-BYTES" {
		t.Errorf("body = %q, want proxied image", body)
	}
}

func TestHandleImageVideoServesThumbnail(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(feed(map[string]string{
		"media_type": "video",
		"hdurl":      base + "/thumb",
		"basic_html": `<iframe src="https://youtube/embed/x"></iframe>`,
	}))

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if body := rr.Body.String(); body != "THUMB-BYTES" {
		t.Errorf("video day should serve the thumbnail, got %q", body)
	}
}

func TestHandleImageOtherReturns404(t *testing.T) {
	base, _, setBody := stubNASA(t)
	setBody(feed(map[string]string{"media_type": "other", "title": "Interactive"}))

	rr := httptest.NewRecorder()
	newService(base).HandleImage()(rr, httptest.NewRequest(http.MethodGet, "/image", nil))

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (no image on 'other' days)", rr.Code)
	}
}

func TestUpstreamErrorReturns502(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dead := l.Addr().String()
	l.Close() // nothing listens here now, so requests are refused

	svc := New(&Config{NASABaseURL: fmt.Sprintf("http://%s/apod", dead), CacheTTL: time.Minute})

	rr := httptest.NewRecorder()
	svc.HandleApod()(rr, httptest.NewRequest(http.MethodGet, "/apod", nil))
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rr.Code)
	}
}
