package repository

import (
	"context"
	"AuthDB/utils"
	"log"
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
		log.Fatalf("Error hashing password: %v", err)
		return nil, err
	}
	HashPassword = hashedPassword
	return &User{Login: login, Email: email, Password: hashedPassword}, nil
}
func GetAllUsers() (users []User, err error) {
	query := `select * from users`
	ctx := context.Background()
	rows, err := Dbpool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Login, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return
}

func GetUserById(ctx context.Context, userID string) (u User, err error) {
	query := `select id, login, email, password from users where id = $1`
	row := Dbpool.QueryRow(ctx, query, userID)
	err = row.Scan(&u.ID, &u.Login, &u.Email, &u.Password)
	return
}

func (u *User) Add(ctx context.Context) (err error) {
	query := `insert into users (login, email, password) values ($1, $2, $3)`
	_, err = Dbpool.Exec(ctx, query, u.Login, u.Email, u.Password)
	return
}

func (u *User) Delete(ctx context.Context, userID string) (err error) {
	query := `delete from users where id = $1`
	_, err = Dbpool.Exec(ctx, query, userID)
	return
}

func (u *User) Update(ctx context.Context) (err error) {
	query := `update users set login = $1, email = $2, password = $3 where id = $4`
	_, err = Dbpool.Exec(ctx, query, u.Login, u.Email, u.Password, u.ID)
	return
}
