package scraper_test

import (
	"net/http"
	"testing"

	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
	"github.com/tobyrushton/fantasy-nba/pkg/scraper"
)

func TestScraper_GetTeams(t *testing.T) {
	c := &fakes.FakeClient{}
	s := scraper.New(c)

	rc := getReadCloser(t, "./testdata/nba_teams.html")
	c.DoReturns(&http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	teams, err := s.GetTeams()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// there are 30 nba teams, so we expect to get 30 back
	if len(teams) != 30 {
		t.Fatalf("expected 30 teams, got %d", len(teams))
	}
}
