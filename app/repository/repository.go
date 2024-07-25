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
	err = r.pool.QueryRow(ctx, `select count(*) from users where login = $1 or email = $2`,
		login, email).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count > 0, nil
}

func (r *Repository) DeleteUserByID(ctx context.Context, id int) error {
	query := `delete from users where id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *Repository) FindUser(ctx context.Context, password string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, login, email, password from users where password = $1`,
		password)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

func (r *Repository) UpdateData(ctx context.Context, query, new, old string) error {
	_, err := r.pool.Exec(ctx, query, new, old)
	return err
}
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, login, email, password from users where email = $1`,
		email)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}
func (r *Repository) FindUserByPassword(ctx context.Context, password string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, login, email, password from users where password = $1`,
		password)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}
func (r *Repository) FindUserByLogin(ctx context.Context, login string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, login, email, password from users where login = $1`,
		login)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}
