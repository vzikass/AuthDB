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
func (r *Repository) Login(ctx context.Context, tx pgx.Tx, username string) (*User, error) {
	query := `SELECT id, username, password, email FROM users WHERE username = $1`
	u := User{}

	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.Password, &u.Email)
	} else {
		err = r.pool.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.Password, &u.Email)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &u, nil
}

// Checking if such a user already exists.
// If user already exists with the same username and email,
// then count will be > 0
func (r *Repository) UserExist(ctx context.Context, tx pgx.Tx, username, email string) (exist bool, err error) {
	var count int
	query := `select count(*) from users where username = $1 or email = $2`

	if tx != nil {
		err = tx.QueryRow(ctx, query, username, email).Scan(&count)
	} else {
		err = r.pool.QueryRow(ctx, query, username, email).Scan(&count)
	}
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetByID(ctx context.Context, tx pgx.Tx, id int) (user User, err error) {
	query := `select * from users where id = $1`

	if tx != nil {
		err = tx.QueryRow(ctx, query, id).Scan(
			&user.ID, &user.Username, &user.Email,
			&user.Password, &user.Role, &user.CreatedAt,
		)
	} else {
		err = r.pool.QueryRow(ctx, query, id).Scan(
			&user.ID, &user.Username, &user.Email,
			&user.Password, &user.Role, &user.CreatedAt,
		)
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
	row := r.pool.QueryRow(ctx, `select id, username, email, password from users where email = $1`,
		email)
	err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

func (r *Repository) FindUserByPassword(ctx context.Context, password string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, username, email, password from users where password = $1`,
		password)
	err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

func (r *Repository) FindUserByLogin(ctx context.Context, username string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select id, username, email, password from users where username = $1`,
		username)
	err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}

func (r *Repository) FindUserByID(ctx context.Context, userID int) (u User, err error) {
	row := r.pool.QueryRow(ctx, "select id, username, role from users where id = $1",
		userID)
	err = row.Scan(&u.ID, &u.Username, &u.Role)
	if err != nil{
		return u, fmt.Errorf("failed to query data: %v", err)
	}
	return u, nil
}