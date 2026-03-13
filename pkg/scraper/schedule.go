package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tobyrushton/fantasy-nba/pkg/models"
)

// ScrapedGame holds the raw scraped data before DB ID resolution.
type ScrapedGame struct {
	NBAID         string
	Season        int
	GameDate      time.Time
	HomeTeamNBAID string
	AwayTeamNBAID string
	HomeScore     *int
	AwayScore     *int
}

func (s *Scraper) GetSchedule(teams []models.Team, season int) ([]ScrapedGame, error) {
	games := make([]ScrapedGame, 0)
	parseErrors := make([]string, 0)

	wg := sync.WaitGroup{}

	// scrape home schedule for each team in parallel, then combine results.
	for _, team := range teams {
		wg.Add(1)
		go func(team models.Team) {
			defer wg.Done()

			homeGames, err := s.scrapeHomeGames(team.NBAID, season)
			if err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("team %s: %v", team.Name, err))
				return
			}
			games = append(games, homeGames...)
		}(team)
	}

	wg.Wait()

	if len(parseErrors) > 0 {
		// Return what we have plus a combined error so callers can decide whether to abort
		return games, fmt.Errorf("scrape completed with %d error(s):\n%s",
			len(parseErrors), strings.Join(parseErrors, "\n"))
	}

	return games, nil
}

func (s *Scraper) scrapeHomeGames(homeTeamNBAID string, season int) ([]ScrapedGame, error) {
	url := fmt.Sprintf("https://www.nba.com/team/%s/schedule", homeTeamNBAID)
	res, err := s.getPage(url)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule page: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing schedule page HTML: %w", err)
	}

	var games []ScrapedGame
	var parseErr error

	doc.Find("tbody.Crom_body__UYOcU tr[data-game-id]").Each(func(_ int, row *goquery.Selection) {
		if parseErr != nil {
			return
		}

		// --- filter: regular season only -----------------------------------
		// The Notes column contains "Week N" for regular season games and
		// "Preseason" for preseason games.
		notes := strings.TrimSpace(row.Find("td.TeamScheduleTable_memo__P2B1e").Text())
		if !strings.Contains(notes, "Week") {
			return
		}

		// --- filter: home games only ---------------------------------------
		location := strings.TrimSpace(row.Find(".TeamScheduleTable_location__3HKVr").Text())
		if location != "vs" {
			return
		}

		// --- game id -------------------------------------------------------
		nbaID, exists := row.Attr("data-game-id")
		if !exists || nbaID == "" {
			return
		}

		// --- game date -----------------------------------------------------
		dateText := strings.TrimSpace(row.Find("td:first-child").Text())
		gameDate, err := parseGameDate(dateText, season)
		if err != nil {
			parseErr = fmt.Errorf("game %s: parsing date %q: %w", nbaID, dateText, err)
			return
		}

		// --- opponent team id ----------------------------------------------
		// The opponent link href is  /team/<nbaID>/
		opponentHref, _ := row.Find(".TeamScheduleTable_oppTeam__h0Anj").Attr("href")
		awayTeamNBAID := extractTeamID(opponentHref)
		if awayTeamNBAID == "" {
			parseErr = fmt.Errorf("game %s: could not extract away team id from %q", nbaID, opponentHref)
			return
		}

		// --- scores (nil when game not yet played) -------------------------
		var homeScore, awayScore *int

		gameStatus, _ := row.Attr("data-game-status")
		if gameStatus == "3" { // completed
			// Status cell text looks like "W103 - 121" or "L117 - 116"
			// The first number is the AWAY (opponent) score, the second is HOME.
			statusText := strings.TrimSpace(row.Find("td:nth-child(3) a span:last-child").Text())
			away, home, err := parseScore(statusText)
			if err != nil {
				parseErr = fmt.Errorf("game %s: parsing score %q: %w", nbaID, statusText, err)
				return
			}
			homeScore = &home
			awayScore = &away
		}

		games = append(games, ScrapedGame{
			NBAID:         nbaID,
			Season:        season,
			GameDate:      gameDate,
			HomeTeamNBAID: homeTeamNBAID,
			AwayTeamNBAID: awayTeamNBAID,
			HomeScore:     homeScore,
			AwayScore:     awayScore,
		})
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return games, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// extractTeamID parses the NBA team ID from a URL path like "/team/1610612738/".
func extractTeamID(href string) string {
	// href: /team/1610612738/celtics  or  /team/1610612738/
	parts := strings.Split(strings.Trim(href, "/"), "/")
	// parts: ["team", "1610612738", "celtics"] or ["team", "1610612738"]
	if len(parts) >= 2 && parts[0] == "team" {
		return parts[1]
	}
	return ""
}

// parseScore parses a string like "103 - 121" and returns (awayScore, homeScore).
// For a home game displayed on the team's own schedule page the convention on
// NBA.com is: <opponent score> - <home team score>.
func parseScore(s string) (away, home int, err error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected score format %q", s)
	}
	away, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("away score: %w", err)
	}
	home, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("home score: %w", err)
	}
	return away, home, nil
}

// parseGameDate converts a date string like "Oct 22" or "Jan 3" to a
// time.Time. Because the schedule page omits the year we derive it from the
// season: months Oct–Dec belong to the start year, Jan–Apr to start year + 1.
func parseGameDate(s string, season int) (time.Time, error) {
	s = strings.TrimSpace(s)
	// time.Parse requires a year; we'll add one based on the month.
	// First parse with a placeholder year to get the month.
	t, err := time.Parse("Jan 2 2006", s+" 2000")
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse %q: %w", s, err)
	}
	year := season
	if t.Month() < time.October { // Jan – Sep → second calendar year of the season
		year = season + 1
	}
	t, err = time.Parse("Jan 2 2006", fmt.Sprintf("%s %d", s, year))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
