package scraper

import (
	"net/http"
	"testing"

	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
	"github.com/tobyrushton/fantasy-nba/pkg/models"
)

func TestScraper_GetPlayers(t *testing.T) {
	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_profile.html")
	c.DoReturns(&http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	teams := []models.Team{
		{
			NBAID:        "1610612743",
			Name:         "Denver Nuggets",
			Abbreviation: "DEN",
		},
	}

	players, err := s.GetPlayers(teams)
	if err != nil {
		t.Fatalf("failed to get players: %v", err)
	}

	// The Nuggets had 18 players on their roster at the time of writing this test.
	if len(players) != 18 {
		t.Fatalf("expected 18 players, got %d", len(players))
	}
}
