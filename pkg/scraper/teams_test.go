package scraper

import (
	"net/http"
	"testing"
)

func TestScraper_GetTeams(t *testing.T) {
	s := New(&http.Client{})

	teams, err := s.GetTeams()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// there are 30 nba teams, so we expect to get 30 back
	if len(teams) != 30 {
		t.Fatalf("expected 30 teams, got %d", len(teams))
	}
}
