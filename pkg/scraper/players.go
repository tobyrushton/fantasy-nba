package scraper

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/tobyrushton/fantasy-nba/pkg/models"
)

// GetPlayers retrieves the list of players from the website.
// Requires a list of teams to scrape.
func (s *Scraper) GetPlayers(teams []models.Team) ([]models.Player, error) {
	wg := sync.WaitGroup{}

	players := make([]models.Player, 0)

	// scrape each team in parallel to speed up the process
	for _, team := range teams {
		wg.Add(1)
		go func() {
			teamPlayers, err := s.getPlayersForTeam(team)
			if err != nil {
				return
			}
			players = append(players, teamPlayers...)
			defer wg.Done()
		}()
	}

	// wait for all goroutines to finish before returning the players
	wg.Wait()
	return players, nil
}

func (s *Scraper) getPlayersForTeam(team models.Team) ([]models.Player, error) {
	players := make([]models.Player, 0)
	res, err := s.getPage(fmt.Sprintf("https://www.nba.com/team/%s/%s", team.NBAID, team.Abbreviation))
	if err != nil {
		return players, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// The roster is inside a <table> whose <thead> contains "Player" and "Pos".
	// We locate it by finding a table that has a "Pos" header column.
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		// Only process the roster table (it has a "Pos" header).
		hasPos := false
		table.Find("thead th").Each(func(_ int, th *goquery.Selection) {
			if strings.TrimSpace(th.Text()) == "Pos" {
				hasPos = true
			}
		})
		if !hasPos {
			return
		}

		// Determine which column index holds "Player" and "Pos".
		colIndex := map[string]int{}
		table.Find("thead th").Each(func(i int, th *goquery.Selection) {
			text := strings.TrimSpace(th.Text())
			colIndex[text] = i
		})

		playerCol, hasPlayer := colIndex["Player"]
		posCol, hasPos2 := colIndex["Pos"]
		if !hasPlayer || !hasPos2 {
			return
		}

		// Iterate body rows.
		table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("td")

			nameCell := cells.Eq(playerCol)
			posCell := cells.Eq(posCol)

			fullName := strings.TrimSpace(nameCell.Text())
			position := strings.TrimSpace(posCell.Text())

			if fullName == "" {
				return
			}

			first, last := splitName(fullName)
			players = append(players, models.Player{
				FirstName: first,
				LastName:  last,
				Position:  position,
			})
		})
	})

	return players, nil
}

// splitName splits "Jayson Tatum" → ("Jayson", "Tatum").
// Handles multi-word last names like "Ron Harper Jr." gracefully:
// everything after the first space is treated as the last name.
func splitName(full string) (first, last string) {
	parts := strings.SplitN(full, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
