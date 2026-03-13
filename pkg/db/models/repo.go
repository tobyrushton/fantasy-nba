package models

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../../fakes/repo.go . Repo
type Repo interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateLeague(ctx context.Context, name string, creatorID int64) (*League, error)
	GetLeagues(ctx context.Context) ([]*League, error)
	GetLeagueByID(ctx context.Context, id int) (*League, error)
	DeleteLeague(ctx context.Context, id int, userID int64) error
	JoinLeague(ctx context.Context, leagueID int, userID int64) error
	CreateRoster(ctx context.Context, leagueID, userID int64, playerIDs []int64) error
	GetRostersByLeagueID(ctx context.Context, leagueID int) ([]*TeamRoster, error)
	UpdateRoster(ctx context.Context, leagueID, userID int64, removePlayerIDs, addPlayerIDs []int64) error
	GetPlayers(ctx context.Context, search, position string) ([]Player, error)
}

type PostgresRepo struct {
	db *bun.DB
}

func NewPostgresRepo(db *bun.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateUser(ctx context.Context, username, passwordHash string) (*User, error) {
	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{Username: username}
	err := r.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresRepo) CreateLeague(ctx context.Context, name string, creatorID int64) (*League, error) {
	league := &League{
		Name:      name,
		CreatorID: creatorID,
	}

	_, err := r.db.NewInsert().Model(league).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return league, nil
}

func (r *PostgresRepo) GetLeagues(ctx context.Context) ([]*League, error) {
	var leagues []*League
	err := r.db.NewSelect().Model(&leagues).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

func (r *PostgresRepo) GetLeagueByID(ctx context.Context, id int) (*League, error) {
	league := &League{ID: int64(id)}
	err := r.db.NewSelect().Model(league).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return league, nil
}

func (r *PostgresRepo) DeleteLeague(ctx context.Context, id int, userID int64) error {
	// Check if league exists and if user is the creator
	league, err := r.GetLeagueByID(ctx, id)
	if err != nil {
		return err
	}
	if league == nil {
		return nil // League not found, treat as successful delete
	}
	if league.CreatorID != userID {
		return nil // User is not the creator, treat as successful delete
	}

	_, err = r.db.NewDelete().Model(&League{ID: int64(id)}).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *PostgresRepo) JoinLeague(ctx context.Context, leagueID int, userID int64) error {
	// Check if league exists
	league, err := r.GetLeagueByID(ctx, leagueID)
	if err != nil {
		return err
	}
	if league == nil {
		return fmt.Errorf("league with id: %d does not exist", leagueID)
	}

	membership := &LeagueMembership{
		LeagueID: int64(leagueID),
		UserID:   userID,
	}

	_, err = r.db.NewInsert().Model(membership).Exec(ctx)
	return err
}

func (r *PostgresRepo) CreateRoster(ctx context.Context, leagueID, userID int64, playerIDs []int64) error {
	// Check if league exists
	league, err := r.GetLeagueByID(ctx, int(leagueID))
	if err != nil {
		return err
	}
	if league == nil {
		return fmt.Errorf("league with id: %d does not exist", leagueID)
	}

	roster := make([]*TeamRoster, len(playerIDs))
	for i, playerID := range playerIDs {
		roster[i] = &TeamRoster{
			LeagueID: leagueID,
			UserID:   userID,
			PlayerID: playerID,
		}
	}

	_, err = r.db.NewInsert().Model(&roster).Exec(ctx)
	return err
}

func (r *PostgresRepo) GetRostersByLeagueID(ctx context.Context, leagueID int) ([]*TeamRoster, error) {
	var rosters []*TeamRoster
	err := r.db.NewSelect().Model(&rosters).Where("league_id = ?", leagueID).Relation("Player").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return rosters, nil
}

func (r *PostgresRepo) UpdateRoster(ctx context.Context, leagueID, userID int64, removePlayerIDs, addPlayerIDs []int64) error {
	// Check if league exists
	league, err := r.GetLeagueByID(ctx, int(leagueID))
	if err != nil {
		return err
	}
	if league == nil {
		return fmt.Errorf("league with id: %d does not exist", leagueID)
	}

	// Remove players from roster
	if len(removePlayerIDs) > 0 {
		_, err = r.db.NewDelete().Model(&TeamRoster{}).
			Where("league_id = ? AND user_id = ? AND player_id IN (?)", leagueID, userID, bun.List(removePlayerIDs)).
			Exec(ctx)
		if err != nil {
			return err
		}
	}

	// Add players to roster
	if len(addPlayerIDs) > 0 {
		newRoster := make([]*TeamRoster, len(addPlayerIDs))
		for i, playerID := range addPlayerIDs {
			newRoster[i] = &TeamRoster{
				LeagueID: leagueID,
				UserID:   userID,
				PlayerID: playerID,
			}
		}

		_, err = r.db.NewInsert().Model(&newRoster).Exec(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepo) GetPlayers(ctx context.Context, search, position string) ([]Player, error) {
	var players []Player
	query := r.db.NewSelect().Model(&players)

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if position != "" {
		query = query.Where("position = ?", position)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

var _ Repo = (*PostgresRepo)(nil)
