package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tobyrushton/fantasy-nba/pkg/config"
	"github.com/tobyrushton/fantasy-nba/pkg/db"
	"github.com/tobyrushton/fantasy-nba/pkg/db/models"
)

func main() {
	cfg := config.MustLoadConfig()

	db, err := db.NewDb(fmt.Sprintf("postgres://admin:%s@localhost:5432/postgres?sslmode=disable", cfg.DB_PASSWORD))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	ctx := context.Background()

	_, err = db.NewCreateTable().
		Model((*models.Team)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create teams table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.Player)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create players table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.Game)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create games table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.PlayerGameStats)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create player_game_stats table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.User)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.League)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create leagues table: %v", err)
	}

	_, err = db.NewCreateTable().
		Model((*models.LeagueMembership)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create league_memberships table: %v", err)
	}

	log.Println("Database migration completed successfully.")
}
