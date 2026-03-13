package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ScrapedTeam struct {
	Name         string
	Abbreviation string
	NBAID        string
}

// GetTeams retrieves the list of NBA teams from basketball-reference
func (s *Scraper) GetTeams() ([]ScrapedTeam, error) {
	res, err := s.getPage("https://www.nba.com/teams")
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

	teams := make([]ScrapedTeam, 0)

	doc.Find("a.NavTeamList_ntlTeam__9K_aX").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("data-text")
		abbr, _ := s.Attr("data-content")

		// Extract ID from the logo img src:
		// e.g. https://cdn.nba.com/logos/nba/1610612738/primary/L/logo.svg
		imgSrc, exists := s.Find("img").Attr("src")
		id := ""
		if exists {
			parts := strings.Split(imgSrc, "/")
			for i, p := range parts {
				if p == "nba" && i+1 < len(parts) {
					id = parts[i+1]
					break
				}
			}
		}

		teams = append(teams, ScrapedTeam{
			NBAID:        id,
			Abbreviation: abbr,
			Name:         name,
		})
	})

	return teams, nil
}
