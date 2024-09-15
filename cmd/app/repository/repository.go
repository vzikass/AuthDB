// If you see tx, then this function is interacting with the transaction
// If you need to use a transaction, then tx should not be nill
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
func (r *Repository) Login(ctx context.Context, tx pgx.Tx, login string) (u User, err error) {
	query := `select id, login, password, email from users where login = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, login).Scan(&u.ID, &u.Login, &u.Password, &u.Email)
	} else {
		err = r.pool.QueryRow(ctx, query, login).Scan(&u.ID, &u.Login, &u.Password, &u.Email)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return u, nil
		}
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

// Checking if such a user already exists.
// If user already exists with the same login and email,
// then count will be > 0
func (r *Repository) UserExist(ctx context.Context, tx pgx.Tx, login, email string) (exist bool, err error) {
	var count int
	query := `select count(*) from users where login = $1 or email = $2`

	if tx != nil {
		err = tx.QueryRow(ctx, query, login, email).Scan(&count)
	} else {
		err = r.pool.QueryRow(ctx, query, login, email).Scan(&count)
	}
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetByID(ctx context.Context, tx pgx.Tx, id int) (user User, err error) {
	query := `select * from users where id = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, id).Scan(&user.ID, &user.Login, &user.Email, &user.Password)
	} else {
		err = r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Login, &user.Email, &user.Password)
	}
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *Repository) DeleteUserByID(ctx context.Context, id int) error {
	query := `delete from users where id = $1`
	
	_, err := r.pool.Exec(ctx, query, id)
	return err
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
