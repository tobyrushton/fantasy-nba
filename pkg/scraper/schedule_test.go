package scraper

import (
	"net/http"
	"testing"

	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
)

func TestScraper_scrapeHomeGames(t *testing.T) {
	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_schedule.html")
	c.DoReturnsOnCall(0, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	homeGames, err := s.scrapeHomeGames("denver-nuggets", 2025)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(homeGames) != 41 {
		t.Fatalf("expected 41 home games, got %d", len(homeGames))
	}
}

func TestScraper_GetHomeGames(t *testing.T) {
	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_schedule.html")
	c.DoReturnsOnCall(0, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	rc = getReadCloser(t, "./testdata/warriors_schedule.html")
	c.DoReturnsOnCall(1, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	teams := []ScrapedTeam{
		{NBAID: "NUGGETS"},
		{NBAID: "WARRIORS"},
	}

	games, err := s.GetSchedule(teams, 2025)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(games) != 82 {
		t.Fatalf("expected 82 games, got %d", len(games))
	}
}
