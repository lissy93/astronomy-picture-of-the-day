// Package shared holds the APOD fetching, caching, and HTTP handlers.
package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kelseyhightower/envconfig"
)

// cacheControl is set on API responses; APOD only changes once a day.
const cacheControl = "public, max-age=900, s-maxage=900"

type Config struct {
	Port               string        `envconfig:"PORT" default:"8080"`
	CORSAllowedOrigins string        `envconfig:"CORS_ALLOWED_ORIGINS" default:"*"`
	NASABaseURL        string        `envconfig:"NASA_BASE_URL" default:"https://science.nasa.gov/wp-json/wp/v2/apod-basic"`
	CacheTTL           time.Duration `envconfig:"CACHE_TTL" default:"15m"`
}

func NewConfig() (*Config, error) {
	var c Config
	if err := envconfig.Process("apod", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Response is the JSON served at /apod, matching the old api.nasa.gov shape.
type Response struct {
	Copyright    string `json:"copyright,omitempty"`
	Date         string `json:"date,omitempty"`
	Explanation  string `json:"explanation,omitempty"`
	HdURL        string `json:"hdurl,omitempty"`
	MediaType    string `json:"media_type,omitempty"`
	Title        string `json:"title,omitempty"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// image returns the still to serve at /image, or "" if the day has none.
func (r *Response) image() string {
	if r.MediaType == "video" {
		return r.ThumbnailURL
	}
	if r.URL != "" {
		return r.URL
	}
	return r.HdURL
}

// Service fetches (and briefly caches) the current APOD, and proxies its image.
type Service struct {
	conf   *Config
	client *http.Client

	mu     sync.Mutex
	cache  *Response
	expiry time.Time
}

func New(conf *Config) *Service {
	return &Service{
		conf:   conf,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

var (
	defaultOnce sync.Once
	defaultSvc  *Service
	defaultErr  error
)

// Default returns a process-wide Service, reused across warm serverless calls.
func Default() (*Service, error) {
	defaultOnce.Do(func() {
		conf, err := NewConfig()
		if err != nil {
			defaultErr = err
			return
		}
		defaultSvc = New(conf)
	})
	return defaultSvc, defaultErr
}

// Routes builds the API router: CORS, gzip, and the /apod and /image endpoints.
func (s *Service) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		cors.Handler(s.corsOptions()),
		middleware.Compress(5),
	)
	r.Get("/apod", s.HandleApod())
	r.Get("/image", s.HandleImage())
	return r
}

// CORS wraps a handler with the configured cross-origin policy.
func (s *Service) CORS(next http.Handler) http.Handler {
	return cors.Handler(s.corsOptions())(next)
}

func (s *Service) corsOptions() cors.Options {
	origins := strings.Split(s.conf.CORSAllowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
		AllowedHeaders: []string{"Accept", "Content-Type"},
		MaxAge:         300,
	}
}

// Fetch returns the latest APOD, from the in-memory cache while it is fresh.
func (s *Service) Fetch(ctx context.Context) (*Response, error) {
	s.mu.Lock()
	if s.cache != nil && time.Now().Before(s.expiry) {
		cached := *s.cache
		s.mu.Unlock()
		return &cached, nil
	}
	s.mu.Unlock()

	endpoint, err := url.Parse(s.conf.NASABaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid NASA base url: %w", err)
	}
	q := endpoint.Query()
	q.Set("per_page", "1")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-apod")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the APOD feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APOD feed returned status %d", resp.StatusCode)
	}

	var posts []feedPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("could not decode the APOD feed: %w", err)
	}
	if len(posts) == 0 {
		return nil, errors.New("APOD feed returned no entries")
	}
	result := posts[0].toResponse()

	s.mu.Lock()
	s.cache = result
	s.expiry = time.Now().Add(s.conf.CacheTTL)
	s.mu.Unlock()

	out := *result
	return &out, nil
}

// HandleApod returns the current APOD metadata as JSON.
func (s *Service) HandleApod() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apod, err := s.Fetch(r.Context())
		if err != nil {
			log.Printf("apod: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if err := json.NewEncoder(w).Encode(apod); err != nil {
			log.Printf("apod: encode: %v", err)
		}
	}
}

// HandleImage proxies the current image, or a video day's thumbnail, from a stable URL.
func (s *Service) HandleImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apod, err := s.Fetch(r.Context())
		if err != nil {
			log.Printf("image: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}

		src := apod.image()
		if src == "" {
			http.Error(w, "no image available today", http.StatusNotFound)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, src, nil)
		if err != nil {
			log.Printf("image: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("image: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		if resp.ContentLength >= 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		}
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("image: copy: %v", err)
		}
	}
}

// StaticHandler serves the embedded frontend with a sensible cache policy.
func StaticHandler(fsys http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(fsys)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fileServer.ServeHTTP(w, r)
	}
}
