package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/tobyrushton/fantasy-nba/pkg/config"
	"github.com/tobyrushton/fantasy-nba/pkg/db"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
	"github.com/tobyrushton/fantasy-nba/pkg/scraper"
	"github.com/uptrace/bun"
)

// This command is in charge of seeding the database.
func main() {
	cfg := config.MustLoadConfig()

	db, err := db.NewDb(fmt.Sprintf("postgres://admin:%s@localhost:5432/postgres?sslmode=disable", cfg.DB_PASSWORD))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// c := client.NewLimitedClient(1000)

	s := scraper.New(&http.Client{})
	ctx := context.Background()

	deleteAll(ctx, db)

	log.Println("Seeding database...")

	log.Println("Scraping teams")

	teams, err := s.GetTeams()
	if err != nil {
		log.Fatalf("failed to scrape teams: %v", err)
	}

	dbTeams := make([]*models.Team, len(teams))
	for i, team := range teams {
		dbTeams[i] = &models.Team{
			Name:         team.Name,
			Abbreviation: team.Abbreviation,
			NBAID:        team.NBAID,
		}
	}
	if _, err := db.NewInsert().Model(&dbTeams).Returning("*").Exec(ctx); err != nil {
		log.Fatalf("failed to insert teams: %v", err)
	}

	log.Println("Scraping players")

	players, err := s.GetPlayers(teams)
	if err != nil {
		log.Fatalf("failed to scrape players: %v", err)
	}

	dbPlayers := make([]*models.Player, len(players))
	for i, player := range players {
		team := findTeam(dbTeams, player.TeamNBAID)
		if team == nil {
			log.Fatalf("failed to find team for player %s %s", player.FirstName, player.LastName)
		}
		dbPlayers[i] = &models.Player{
			TeamID:    team.ID,
			FirstName: player.FirstName,
			LastName:  player.LastName,
			Position:  player.Position,
			NBAID:     player.NBAID,
		}
	}
	if _, err := db.NewInsert().Model(&dbPlayers).Returning("*").Exec(ctx); err != nil {
		log.Fatalf("failed to insert players: %v", err)
	}

	log.Println("Scraping games")

	games, err := s.GetSchedule(teams, 2025)
	if err != nil {
		log.Fatalf("failed to scrape games: %v", err)
	}

	dbGames := make([]*models.Game, len(games))
	for i, game := range games {
		homeTeam := findTeam(dbTeams, game.HomeTeamNBAID)
		if homeTeam == nil {
			log.Fatalf("failed to find home team for game %s", game.NBAID)
		}
		awayTeam := findTeam(dbTeams, game.AwayTeamNBAID)
		if awayTeam == nil {
			log.Fatalf("failed to find away team for game %s", game.NBAID)
		}
		dbGames[i] = &models.Game{
			NBAID:      game.NBAID,
			Season:     game.Season,
			GameDate:   game.GameDate,
			HomeTeamID: homeTeam.ID,
			AwayTeamID: awayTeam.ID,
		}
	}
	if _, err := db.NewInsert().Model(&dbGames).Returning("*").Exec(ctx); err != nil {
		log.Fatalf("failed to insert games: %v", err)
	}
	if err := db.NewSelect().Model(&dbGames).Relation("HomeTeam").Relation("AwayTeam").Scan(ctx); err != nil {
		log.Fatalf("failed to select games with teams: %v", err)
	}

	log.Println("Scraping game stats")

	gamesToScrape := make([]string, 0)

	for _, game := range dbGames {
		if game.GameDate.Before(time.Now()) {
			gamesToScrape = append(gamesToScrape, fmt.Sprintf("%s-vs-%s-%s", game.HomeTeam.Abbreviation, game.AwayTeam.Abbreviation, game.NBAID))
		}
	}

	playerGameStats, err := s.GetGameStats(gamesToScrape)
	if err != nil {
		log.Fatalf("failed to scrape game stats: %v", err)
	}

	dbPlayerGameStats := make([]*models.PlayerGameStats, len(playerGameStats))

	for i, stat := range playerGameStats {
		player := findPlayer(dbPlayers, strconv.Itoa(int(stat.PlayerID)))
		if player == nil {
			// This can happen for players who are inactive and don't appear in the stats page, so we log and skip them.
			log.Printf("failed to find player for game stat %s", strconv.Itoa(int(stat.PlayerID)))
			continue
		}
		game := findGame(dbGames, stat.GameID)
		if game == nil {
			log.Fatalf("failed to find game for game stat with game NBAID %s", stat.GameID)
		}
		dbPlayerGameStats[i] = &models.PlayerGameStats{
			PlayerID:   player.ID,
			GameID:     game.ID,
			TeamID:     player.TeamID,
			DidNotPlay: stat.DidNotPlay,
			Points:     stat.Points,
			Rebounds:   stat.Rebounds,
			Assists:    stat.Assists,
			Steals:     stat.Steals,
		}
	}
	log.Println("Successfully seeded database")
}

func deleteAll(ctx context.Context, db *bun.DB) {
	log.Println("Erasing database...")
	if _, err := db.NewDelete().Model((*models.PlayerGameStats)(nil)).Where("1=1").Exec(ctx); err != nil {
		log.Fatalf("failed to erase database: %v", err)
	}
	if _, err := db.NewDelete().Model((*models.Game)(nil)).Where("1=1").Exec(ctx); err != nil {
		log.Fatalf("failed to erase database: %v", err)
	}
	if _, err := db.NewDelete().Model((*models.Player)(nil)).Where("1=1").Exec(ctx); err != nil {
		log.Fatalf("failed to erase database: %v", err)
	}
	if _, err := db.NewDelete().Model((*models.Team)(nil)).Where("1=1").Exec(ctx); err != nil {
		log.Fatalf("failed to erase database: %v", err)
	}
	log.Println("Successfully erased db")
}

func findTeam(teams []*models.Team, nbaID string) *models.Team {
	for _, team := range teams {
		if team.NBAID == nbaID {
			return team
		}
	}
	return nil
}

func findPlayer(players []*models.Player, nbaID string) *models.Player {
	for _, player := range players {
		if player.NBAID == nbaID {
			return player
		}
	}
	return nil
}

func findGame(games []*models.Game, nbaID string) *models.Game {
	for _, game := range games {
		if game.NBAID == nbaID {
			return game
		}
	}
	return nil
}
