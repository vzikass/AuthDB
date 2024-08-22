package repository

import (
	"AuthDB/utils"
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Email    string `json:"email" db:"email"`
}

var (
	HashPassword string
)

func NewUser(login, email, password string) (*User, error) {
	hashedPassword, err := utils.GenerateHash(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	HashPassword = hashedPassword
	user := &User{
		Login:    login,
		Email:    email,
		Password: hashedPassword,
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
		if err := rows.Scan(&user.ID, &user.Login, &user.Email, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserById(ctx context.Context, userID int) (u User, err error) {
	query := `select id, login, email, password from users where id = $1`
	row := Dbpool.QueryRow(ctx, query, userID)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	return
}

func (u *User) Add(ctx context.Context, tx pgx.Tx) (err error) {
	if tx != nil {
		_, err := tx.Exec(ctx, "INSERT INTO users (login, email, password) VALUES ($1, $2, $3)", u.Login, u.Email, u.Password)
		return err
	} else {
		_, err := Dbpool.Exec(ctx, "INSERT INTO users (login, email, password) VALUES ($1, $2, $3)", u.Login, u.Email, u.Password)
		return err
	}
}

func (u *User) Delete(ctx context.Context, tx pgx.Tx, userID int) (err error) {
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

func (u *User) Update(ctx context.Context, tx pgx.Tx) (err error) {
	query := `update users set login = $1, email = $2, password = $3 where id = $4`
	if tx != nil {
		_, err = tx.Exec(ctx, query, u.Login, u.Email, u.Password, u.ID)
	} else {
		_, err = Dbpool.Exec(ctx, query, u.Login, u.Email, u.Password, u.ID)
	}
	if err != nil {
		return err
	}
	return nil
}
