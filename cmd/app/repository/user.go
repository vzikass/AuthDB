package repository

import (
	"AuthDB/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int        `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Password  string     `json:"password" db:"password"`
	Email     string     `json:"email" db:"email"`
	Role      string     `json:"role" db:"role"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
}

var (
	HashPassword string
	Role         = "user"
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
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      Role,
		CreatedAt: &curTime,
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
		var createdAt sql.NullTime
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			user.CreatedAt = &createdAt.Time
		} else {
			user.CreatedAt = nil
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserById(ctx context.Context, userID string) (u User, err error) {
	query := `select id, username, email, password from users where id = $1`
	row := Dbpool.QueryRow(ctx, query, userID)
	err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	return
}

func (u *User) Add(ctx context.Context, tx pgx.Tx) (err error) {
	if tx != nil {
		_, err := tx.Exec(ctx, "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
			u.Username, u.Email, u.Password)
		return err
	} else {
		_, err := Dbpool.Exec(ctx, "INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
			u.Username, u.Email, u.Password)
		return err
	}
}

func (u *User) AddAdminUser(ctx context.Context, pool *pgxpool.Pool, tx pgx.Tx) error {
	var exists bool

	// check if user exists
	err := Dbpool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM goadmin_users WHERE username = $1
		);
	`, u.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error when checking admin user exists: %v", err)
	}

	if exists {
		log.Printf("User '%s' already exists, skip creating.\n", u.Username)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// User creating
	var userID int
	err = pool.QueryRow(ctx, `
		INSERT INTO goadmin_users (username, password, role_id)
		VALUES ($1, $2, (SELECT id FROM goadmin_roles WHERE slug = 'admin'))
		RETURNING id
	`, u.Username, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return fmt.Errorf("error creating admin user: %w", err)
	}
	return nil
}

// func getQueryRow(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) pgx.Row {
// 	if tx != nil {
// 		return tx.QueryRow(ctx, query, args...)
// 	}
// 	return Dbpool.QueryRow(ctx, query, args...)
// }

// func getExec(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) error {
// 	var err error

// 	if tx != nil {
// 		_, err = tx.Exec(ctx, query, args...)
// 		if err != nil {
// 			return fmt.Errorf("transaction execution error: %v", err)
// 		}
// 	} else {
// 		_, err = Dbpool.Exec(ctx, query, args...)
// 		if err != nil {
// 			return fmt.Errorf("execution error: %v", err)
// 		}
// 	}

// 	return nil
// }

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
	query := `update users set username = $1, email = $2, password = $3 where id = $4`
	if tx != nil {
		_, err = tx.Exec(ctx, query, u.Username, u.Email, u.Password, u.ID)
	} else {
		_, err = Dbpool.Exec(ctx, query, u.Username, u.Email, u.Password, u.ID)
	}
	if err != nil {
		return err
	}
	return nil
}
