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

var _ Repo = (*PostgresRepo)(nil)
