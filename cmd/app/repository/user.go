package repository

import (
	"AuthDB/utils"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Email    string `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

var (
	HashPassword string
)

// Creating new user.
// User struct receives hashed password
func NewUser(username, email, password string) (*User, error) {
	hashedPassword, err := utils.GenerateHash(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	curTime := time.Now()
	HashPassword = hashedPassword
	user := &User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		CreatedAt: curTime,
	}
	return user, nil
}

func GetAllUsers(ctx context.Context, tx pgx.Tx) ([]User, error) {
	var err error
	var users []User
	var rows pgx.Rows

	if tx != nil {
		rows, err = tx.Query(ctx, "select * from users")
	} else {
		rows, err = Dbpool.Query(ctx, "select * from users")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserById(ctx context.Context, userID string) (u User, err error) {
	query := `select id, username, email, password, created_at from users where id = $1`
	row := Dbpool.QueryRow(ctx, query, userID)
	err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt)
	return
}

func (u *User) Add(ctx context.Context, tx pgx.Tx) (err error) {
	if tx != nil {
		_, err := tx.Exec(ctx, "INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, $4)",
		 u.Username, u.Email, u.Password, u.CreatedAt)
		return err
	} else {
		_, err := Dbpool.Exec(ctx, "INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, $4)",
		 u.Username, u.Email, u.Password, u.CreatedAt)
		return err
	}
}

func (u *User) DeleteByID(ctx context.Context, tx pgx.Tx, userID int) (err error) {
	query := `delete from users where id = $1`
	if tx != nil {
		_, err = tx.Exec(ctx, query, userID)
	} else {
		_, err = Dbpool.Exec(ctx, query, userID)
	}
	if err != nil {
		return err
	}
	return nil
}

func (u *User) UpdateByID(ctx context.Context, tx pgx.Tx) (err error) {
	query := `update users set username = $1, email = $2, password = $3, created_at = $4 where id = $4`
	if tx != nil {
		_, err = tx.Exec(ctx, query, u.Username, u.Email, u.Password, u.CreatedAt, u.ID)
	} else {
		_, err = Dbpool.Exec(ctx, query, u.Username, u.Email, u.Password, u.CreatedAt, u.ID)
	}
	if err != nil {
		return err
	}
	return nil
}
