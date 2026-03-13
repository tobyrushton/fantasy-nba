// Package models contains bun ORM models for the fantasy NBA database.
package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Team represents an NBA franchise.
type Team struct {
	bun.BaseModel `bun:"table:teams,alias:t"`

	ID           int64     `bun:"id,pk,autoincrement"`
	NBAID        string    `bun:"nba_id,notnull,unique"`
	Name         string    `bun:"name,notnull"`
	Abbreviation string    `bun:"abbreviation,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// Player represents an NBA player.
type Player struct {
	bun.BaseModel `bun:"table:players,alias:p"`

	ID        int64     `bun:"id,pk,autoincrement"`
	NBAID     string    `bun:"nba_id,notnull,unique"`
	FirstName string    `bun:"first_name,notnull"`
	LastName  string    `bun:"last_name,notnull"`
	Position  string    `bun:"position,notnull"` // PG, SG, SF, PF, C
	TeamID    int64     `bun:"team_id,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Team *Team `bun:"rel:belongs-to,join:team_id=id"`
}

// Game represents a single NBA game.
type Game struct {
	bun.BaseModel `bun:"table:games,alias:g"`

	ID         int64     `bun:"id,pk,autoincrement"`
	NBAID      string    `bun:"nba_id,notnull,unique"`
	Season     int       `bun:"season,notnull"`
	GameDate   time.Time `bun:"game_date,notnull"`
	HomeTeamID int64     `bun:"home_team_id,notnull"`
	AwayTeamID int64     `bun:"away_team_id,notnull"`
	HomeScore  *int      `bun:"home_score"`
	AwayScore  *int      `bun:"away_score"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	HomeTeam *Team `bun:"rel:belongs-to,join:home_team_id=id"`
	AwayTeam *Team `bun:"rel:belongs-to,join:away_team_id=id"`
}

// PlayerGameStats holds the per-game statistics for a player needed to
// calculate fantasy points: points, rebounds, assists, steals, blocks,
// turnovers, and three-pointers made.
type PlayerGameStats struct {
	bun.BaseModel `bun:"table:player_game_stats,alias:pgs"`

	ID                int64     `bun:"id,pk,autoincrement"`
	PlayerID          int64     `bun:"player_id,notnull"`
	GameID            int64     `bun:"game_id,notnull"`
	TeamID            int64     `bun:"team_id,notnull"`
	DidNotPlay        bool      `bun:"did_not_play,notnull,default:false"`
	Points            int       `bun:"points,notnull,default:0"`
	Rebounds          int       `bun:"rebounds,notnull,default:0"`
	Assists           int       `bun:"assists,notnull,default:0"`
	Steals            int       `bun:"steals,notnull,default:0"`
	Blocks            int       `bun:"blocks,notnull,default:0"`
	Turnovers         int       `bun:"turnovers,notnull,default:0"`
	ThreePointersMade int       `bun:"three_pointers_made,notnull,default:0"`
	FreeThrowsMade    int       `bun:"free_throws_made,notnull,default:0"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt         time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Player *Player `bun:"rel:belongs-to,join:player_id=id"`
	Game   *Game   `bun:"rel:belongs-to,join:game_id=id"`
	Team   *Team   `bun:"rel:belongs-to,join:team_id=id"`
}
