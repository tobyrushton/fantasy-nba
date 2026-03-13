package scraper

import (
	"net/http"
	"testing"

	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
)

func TestScraper_scrapeGameStats(t *testing.T) {
	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_warriors_stats.html")
	c.DoReturnsOnCall(0, &http.Response{
		StatusCode: 200,
		Body:       rc,
	}, nil)

	stats, err := s.scrapeGameStats("den-vs-gsw-23499485")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats) != 34 {
		t.Errorf("expected 34 stats, got %d", len(stats))
	}
}
