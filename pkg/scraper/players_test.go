package scraper

import (
	"net/http"
	"testing"

	"github.com/tobyrushton/fantasy-nba/pkg/fakes"
)

func TestScraper_GetPlayers(t *testing.T) {
	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_profile.html")
	c.DoReturnsOnCall(0, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	rc = getReadCloser(t, "./testdata/warriors_profile.html")
	c.DoReturnsOnCall(1, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	teams := []ScrapedTeam{
		{
			NBAID:        "1610612743",
			Name:         "Denver Nuggets",
			Abbreviation: "DEN",
		},
		{
			NBAID:        "1610612744",
			Name:         "Golden State Warriors",
			Abbreviation: "GSW",
		},
	}

	players, err := s.GetPlayers(teams)
	if err != nil {
		t.Fatalf("failed to get players: %v", err)
	}

	// The Nuggets had 18 players on their roster at the time of writing this test.
	// The Warriors have 17 so in total 35 players should be found.
	if len(players) != 35 {
		t.Fatalf("expected 35 players, got %d", len(players))
	}
}

func TestScraper_getPlayersForTeam(t *testing.T) {
	team := ScrapedTeam{
		NBAID:        "1610612743",
		Name:         "Denver Nuggets",
		Abbreviation: "DEN",
	}

	c := &fakes.FakeClient{}
	s := New(c)

	rc := getReadCloser(t, "./testdata/nuggets_profile.html")
	c.DoReturnsOnCall(0, &http.Response{
		StatusCode: http.StatusOK,
		Body:       rc,
	}, nil)

	players, err := s.getPlayersForTeam(team)
	if err != nil {
		t.Fatalf("failed to get players: %v", err)
	}

	// The Nuggets had 18 players on their roster at the time of writing this test.
	if len(players) != 18 {
		t.Fatalf("expected 18 players, got %d", len(players))
	}
}
