package scraper

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type ScrapedPlayerStats struct {
	PlayerID          int64
	GameID            string
	TeamID            int64
	DidNotPlay        bool
	Points            int
	Rebounds          int
	Assists           int
	Steals            int
	Blocks            int
	Turnovers         int
	ThreePointersMade int
	FreeThrowsMade    int
}

// --- JSON shape of __NEXT_DATA__ (only fields we need) ---

type nextData struct {
	Props struct {
		PageProps struct {
			Game struct {
				GameID   string   `json:"gameId"`
				HomeTeam teamData `json:"homeTeam"`
				AwayTeam teamData `json:"awayTeam"`
			} `json:"game"`
		} `json:"pageProps"`
	} `json:"props"`
}

type teamData struct {
	TeamID   int64        `json:"teamId"`
	Players  []playerData `json:"players"`
	Inactive []struct {
		PersonID int64 `json:"personId"`
	} `json:"inactives"`
}

type playerData struct {
	PersonID   int64       `json:"personId"`
	Comment    string      `json:"comment"` // "DNP - Coach's Decision" etc.
	Statistics playerStats `json:"statistics"`
}

type playerStats struct {
	Points            int `json:"points"`
	ReboundsTotal     int `json:"reboundsTotal"`
	Assists           int `json:"assists"`
	Steals            int `json:"steals"`
	Blocks            int `json:"blocks"`
	Turnovers         int `json:"turnovers"`
	ThreePointersMade int `json:"threePointersMade"`
	FreeThrowsMade    int `json:"freeThrowsMade"`
}

func (s *Scraper) GetGameStats(gameIDs []string) ([]ScrapedPlayerStats, error) {
	allStats := make([]ScrapedPlayerStats, 0)
	wg := sync.WaitGroup{}

	for _, gameID := range gameIDs {
		wg.Add(1)
		go func(gID string) {
			defer wg.Done()
			stats, err := s.scrapeGameStats(gID)
			if err != nil {
				return
			}
			allStats = append(allStats, stats...)
		}(gameID)
	}

	wg.Wait()
	return allStats, nil
}

func (s *Scraper) scrapeGameStats(gameID string) ([]ScrapedPlayerStats, error) {
	url := fmt.Sprintf("https://www.nba.com/game/%s/box-score", gameID)
	res, err := s.getPage(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	results := make([]ScrapedPlayerStats, 0)

	raw := doc.Find("script#__NEXT_DATA__").Text()
	if raw == "" {
		return nil, fmt.Errorf("__NEXT_DATA__ script tag not found")
	}

	var nd nextData
	if err := json.Unmarshal([]byte(raw), &nd); err != nil {
		return nil, fmt.Errorf("failed to parse __NEXT_DATA__: %w", err)
	}

	game := nd.Props.PageProps.Game

	// Build a set of inactive personIDs per team for quick lookup.
	inactiveHome := inactiveSet(game.HomeTeam.Inactive)
	inactiveAway := inactiveSet(game.AwayTeam.Inactive)

	results = append(results, extractTeam(game.HomeTeam, gameID, inactiveHome)...)
	results = append(results, extractTeam(game.AwayTeam, gameID, inactiveAway)...)

	if len(results) == 0 {
		return nil, fmt.Errorf("no player stats found in __NEXT_DATA__")
	}

	return results, nil
}

func extractTeam(team teamData, gameID string, inactives map[int64]bool) []ScrapedPlayerStats {
	var out []ScrapedPlayerStats
	gameNBAID := strings.Split(gameID, "-")[3] // Extract NBAID from gameID

	for _, p := range team.Players {
		dnp := inactives[p.PersonID] || isDNP(p.Comment)
		s := ScrapedPlayerStats{
			PlayerID:   p.PersonID,
			GameID:     gameNBAID,
			TeamID:     team.TeamID,
			DidNotPlay: dnp,
		}
		if !dnp {
			s.Points = p.Statistics.Points
			s.Rebounds = p.Statistics.ReboundsTotal
			s.Assists = p.Statistics.Assists
			s.Steals = p.Statistics.Steals
			s.Blocks = p.Statistics.Blocks
			s.Turnovers = p.Statistics.Turnovers
			s.ThreePointersMade = p.Statistics.ThreePointersMade
			s.FreeThrowsMade = p.Statistics.FreeThrowsMade
		}
		out = append(out, s)
	}

	// Also add pure inactives that may not appear in the players list.
	for id := range inactives {
		found := false
		for _, p := range team.Players {
			if p.PersonID == id {
				found = true
				break
			}
		}
		if !found {
			out = append(out, ScrapedPlayerStats{
				PlayerID:   id,
				GameID:     gameNBAID,
				TeamID:     team.TeamID,
				DidNotPlay: true,
			})
		}
	}

	return out
}

// inactiveSet builds a set from the inactives slice for O(1) lookup.
func inactiveSet(inactives []struct {
	PersonID int64 `json:"personId"`
}) map[int64]bool {
	m := make(map[int64]bool, len(inactives))
	for _, p := range inactives {
		m[p.PersonID] = true
	}
	return m
}

// isDNP returns true if the player comment indicates they did not play.
func isDNP(comment string) bool {
	c := strings.ToUpper(comment)
	return strings.HasPrefix(c, "DNP") || strings.HasPrefix(c, "DND") || strings.HasPrefix(c, "NWT")
}
