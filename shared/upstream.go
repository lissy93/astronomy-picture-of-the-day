package shared

import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

var (
	lineBreak   = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlTag     = regexp.MustCompile(`<[^>]*>`)
	spaceBefore = regexp.MustCompile(` ([.,;:!?)])`)
	footer      = regexp.MustCompile(`(?i)(?:<br\s*/?>\s*)+<(?:strong|b)\b`)
	creditLabel = regexp.MustCompile(`(?i)^(?:(?:image|video) )?credit(?: (?:&|and) copyright)?:\s*`)
	mediaSrc    = regexp.MustCompile(`(?i)<(?:source|video|iframe)\b[^>]*\ssrc=["']([^"']+)`)
)

// feedPost is one entry from NASA's apod-basic feed, where text fields are HTML.
type feedPost struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	MediaType   string `json:"media_type"`
	Explanation string `json:"explanation"`
	Copyright   string `json:"copyright"`
	HdURL       string `json:"hdurl"`
	BasicHTML   string `json:"basic_html"`
}

// toResponse maps a feed entry onto the response shape of the old APOD API.
func (p *feedPost) toResponse() *Response {
	r := &Response{
		Date:        p.Date,
		Title:       htmlToText(p.Title),
		Explanation: cleanExplanation(p.Explanation),
		Copyright:   cleanCredit(p.Copyright),
		MediaType:   p.MediaType,
	}
	if p.MediaType == "video" {
		r.URL = videoURL(p.BasicHTML)
		r.ThumbnailURL = standardRes(p.HdURL) // hdurl is a still frame on video days
	} else if p.HdURL != "" {
		r.URL = standardRes(p.HdURL)
		r.HdURL = p.HdURL
	}
	return r
}

// htmlToText strips tags and entities, and closes gaps like "<a>word</a> .".
func htmlToText(s string) string {
	s = lineBreak.ReplaceAllString(s, " ")
	s = html.UnescapeString(htmlTag.ReplaceAllString(s, ""))
	s = strings.Join(strings.Fields(s), " ")
	s = spaceBefore.ReplaceAllString(s, "$1")
	return strings.ReplaceAll(s, "( ", "(")
}

// cleanExplanation drops the "Explanation:" label and footers like "Tomorrow's picture".
func cleanExplanation(s string) string {
	if loc := footer.FindStringIndex(s); loc != nil {
		s = s[:loc[0]]
	}
	return strings.TrimSpace(strings.TrimPrefix(htmlToText(s), "Explanation:"))
}

// cleanCredit drops labels like "Image Credit & Copyright:".
func cleanCredit(s string) string {
	return creditLabel.ReplaceAllString(htmlToText(s), "")
}

// videoURL returns the first embedded video (mp4 or YouTube) in a post's HTML.
func videoURL(page string) string {
	m := mediaSrc.FindStringSubmatch(page)
	if m == nil {
		return ""
	}
	src := html.UnescapeString(m[1])
	if strings.HasPrefix(src, "//") {
		src = "https:" + src
	}
	return src
}

// standardRes strips resize params from a dynamicimage URL, giving NASA's ~1280px default.
func standardRes(hd string) string {
	u, err := url.Parse(hd)
	if err != nil || !strings.Contains(u.Path, "/dynamicimage/") {
		return hd
	}
	u.RawQuery = ""
	return u.String()
}
