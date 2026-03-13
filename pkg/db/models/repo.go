package models

import (
	"context"

	"github.com/uptrace/bun"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../../fakes/repo.go . Repo
type Repo interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateLeague(ctx context.Context, name string, creatorID int64) (*League, error)
	GetLeagues(ctx context.Context) ([]*League, error)
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

var _ Repo = (*PostgresRepo)(nil)
