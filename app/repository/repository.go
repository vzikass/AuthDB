package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Login(ctx context.Context, login string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, login, password, email from users where login = $1`,
		login)
	err = row.Scan(&u.ID, &u.Login, &u.Password, &u.Email)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

func (r *Repository) UserExist(ctx context.Context, login, email string) (_ bool, err error) {
	var count int
	err = r.pool.QueryRow(ctx, "select count(*) from users where login = $1 or email = $2",
		login, email).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count > 0, nil
}
