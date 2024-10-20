// the main logic with all endpoints can be found here

package controller

import (
	"AuthDB/cmd/app/controller/helper"
	"AuthDB/cmd/app/repository"
	"AuthDB/cmd/internal/kafka"
	"AuthDB/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/yandex"
)

type App struct {
	ctx     context.Context
	repo    *repository.Repository
	cache   map[string]repository.User
	cacheMu sync.Mutex
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *App {
	return &App{ctx: ctx, repo: repository.NewRepository(dbpool),
		cache: make(map[string]repository.User)}
}

func (a *App) Routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))

	r.GET("/", a.authorized(a.HomePage))

	r.POST("/signup", a.Signup)
	r.GET("/signup", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupPage(w, "")
	})
	r.POST("/login", a.Login)
	r.GET("/login", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPage(w, "")
	})
	r.POST("/delete", a.authorized(a.DeleteAccount))
	r.GET("/delete", a.authorized(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RenderDeleteConfirmationPage(w)
	}))
	r.POST("/update", a.authorized(a.UpdateData))
	r.GET("/update", a.authorized(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdateUserPage(w, "")
	}))
	r.GET("/logout", a.authorized(a.Logout))

	// View all users (front by bootstrap)
	r.GET("/users", a.authorized(GetAllUsers))

	r.POST("/auth/:provider/callback", a.authCallbackHandlerForRouter)
	r.GET("/auth/:provider", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		gothic.BeginAuthHandler(w, r)
	})
}

func (a *App) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		a.LoginPage(w, "You must provide a username and password")
		return
	}

	user, err := a.repo.Login(a.ctx, nil, username)
	if err != nil {
		a.LoginPage(w, "User not found")
		return
	}

	// Compare user password and login password using byte
	if !utils.CompareHashPassword(password, user.Password) {
		a.LoginPage(w, "Incorrect password")
		return
	}

	// Generate JWT-token with user username
	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		log.Fatalf("Error generate token: %v", err)
		return
	}

	// to protect access to hash
	a.cacheMu.Lock()
	a.cache[token] = user
	a.cacheMu.Unlock()

	// creating cookies with check button remember me
	rememberMe := r.FormValue("remember_me") == "on"
	var livingTime time.Duration
	// if true, the cookie will be kept for 15 days
	// else 1 hour
	if rememberMe {
		livingTime = 24 * time.Hour * 15
	} else {
		livingTime = 1 * time.Hour
	}
	// remember livingTime
	expiration := time.Now().Add(livingTime)
	// Create cookie
	cookie := http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(token),
		Expires:  expiration,
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	// Creating kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
		"event": "login",
		"user_id": "%d",
		"email": "%s",
		"timestamp": "%s"
		}`, user.ID, user.Email, time.Now().UTC().Format(time.RFC3339))),
	}
	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Signup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user := repository.User{}
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	repassword := strings.TrimSpace(r.FormValue("repassword"))

	if username == "" || email == "" || password == "" || repassword == "" {
		a.SignupPage(w, "Not all fields are filled in")
		return
	}

	if password != repassword {
		a.SignupPage(w, "Password mismatch")
		return
	}

	if !helper.IsValidPassword(password) {
		a.SignupPage(w, "The password should not contain only numbers or letters")
		return
	}

	if len(username) <= 4 {
		a.SignupPage(w, "Minimum username length - 4 characters")
		return
	}

	userExist, err := a.repo.UserExist(a.ctx, nil, username, email)
	if err != nil {
		a.SignupPage(w, "Error checking existing user")
		return
	}

	if userExist {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// create a new user and add it to the database in goroutine
	// errors are written to the channel
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		user, err := repository.NewUser(username, email, password)
		if err != nil {
			errCh <- err
			return
		}
		err = user.Add(a.ctx, nil)
		if err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	// read from channel
	// make sure err == nil, if it is, a kafka message is created
	err = <-errCh
	if err != nil {
		a.SignupPage(w, err.Error())
		return
	}
	// Create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "signup",
			"user_id": "%d",
			"email": "%s",
			"timestamp": "%s"
		}`, user.ID, user.Email, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}
	a.LoginPage(w, fmt.Sprintln("Successful signup!"))
}

// A simple function to delete a user's cookie
// To log him out
func (a *App) Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Read cookie
func ReadCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		return value, err
	}

	// Read cookie by name
	// Returns a cookie or gives an error if no cookie with this name is found
	cookie, err := r.Cookie(name)
	if err != nil {
		return value, err
	}
	// cookie string
	str := cookie.Value
	// decode the string
	value, err = url.QueryUnescape(str)
	if err != nil {
		return value, err
	}
	return value, err
}

// check user authorization
func (a *App) authorized(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// read cookie, if err != nil, user is not authorized
		// so redirect it to /login
		token, err := ReadCookie("token", r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// read token
		// lock access to the cache while we work with it
		a.cacheMu.Lock()
		_, ok := a.cache[token]
		a.cacheMu.Unlock()
		// if ok == false (token not found)
		// redirect to login
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// if ok == true (token found)
		// continue processing the request
		next(w, r, p)
	}
}

// delete account with user id
func (a *App) DeleteAccount(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// read cookie
	token, err := ReadCookie("token", r)
	if err != nil {
		log.Printf("Error reading token cookie: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// cache lookup
	user, ok := a.cache[token]
	// if not found redirect to login
	if !ok {
		log.Printf("Token not found in cache")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// if found delete user by id
	err = a.repo.DeleteUserByID(a.ctx, user.ID)
	if err != nil {
		log.Printf("Error deleting user by ID: %v", err)
		http.Error(w, "Something went wrong, please try later", http.StatusInternalServerError)
		return
	}
	// delete token from cache (map)
	delete(a.cache, token)

	// delete cookie to logout
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
	// Create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
		"event": "delete_account",
		"deleteduser_id": "%d",
		"timestamp": "%s",
		}`, user.ID, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// The following 3 functions are the same
// only they update different data and queries to the database
// also kafka messages are created
func (a *App) UpdateUsername(w http.ResponseWriter, oldusername, newusername string) error {
	user, err := a.repo.FindUserByLogin(a.ctx, oldusername)
	if err != nil {
		a.UpdateUserPage(w, "User not found")
		return err
	}
	// old and new username must not be the same
	if user.Username != newusername {
		// set a new username if they do not match
		query := `UPDATE users SET username = $1 WHERE username = $2`
		err = a.repo.UpdateData(a.ctx, query, newusername, oldusername)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil
	}
	// create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "update_password",
			"user_id": "%d",
			"newusername": "%s",
			"timestamp": "%s"
		}`, user.ID, newusername, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}

	a.UpdateUserPage(w, "This username already exists")
	return fmt.Errorf("username already exists")
}

func (a *App) UpdateEmail(w http.ResponseWriter, oldEmail, newEmail string) error {
	user, err := a.repo.FindUserByEmail(a.ctx, oldEmail)
	if err != nil {
		a.UpdateUserPage(w, "User not found")
		return err
	}
	// old and new email must not be the same
	if user.Email != newEmail {
		// set new email
		query := `UPDATE users SET email = $1 WHERE email = $2`
		err = a.repo.UpdateData(a.ctx, query, newEmail, oldEmail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil
	}
	a.UpdateUserPage(w, "This email already exists")
	// create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "update_password",
			"user_id": "%d",
			"newemail": "%s",
			"timestamp": "%s"
		}`, user.ID, newEmail, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}
	return fmt.Errorf("email already exists")
}

func (a *App) UpdatePassword(w http.ResponseWriter, r *http.Request, newPassword string) error {
	userHashedPass := repository.HashPassword
	if len(newPassword) <= 3 {
		a.UpdateUserPage(w, "Minimum field length - 4 characters")
	}
	// user search by old hashed password
	user, err := a.repo.FindUserByPassword(a.ctx, userHashedPass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	// generate new hashed password
	hashedNewPassword, err := utils.GenerateHash(newPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	// old password (not hashed) and new password must not be the same
	if user.Password != newPassword {
		// set net password
		query := `UPDATE users SET password = $1 WHERE password = $2`
		err = a.repo.UpdateData(a.ctx, query, hashedNewPassword, userHashedPass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	a.UpdateUserPage(w, "Passwords need to be different")
	user.Password = hashedNewPassword
	// create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "update_password",
			"user_id": "%d",
			"newpassword": "%s",
			"timestamp": "%s"
		}`, user.ID, newPassword, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}
	return err
}

func (a *App) UpdateData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user := repository.User{}
	// read lines
	oldUsername := r.FormValue("oldUsername")
	newUsername := r.FormValue("newUsername")
	oldEmail := r.FormValue("oldEmail")
	newEmail := r.FormValue("newEmail")
	newPassword := r.FormValue("newPassword")

	// will work a case where the string != ""
	// then update data
	if oldUsername != "" && newUsername != "" {
		err := a.UpdateUsername(w, oldUsername, newUsername)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if oldEmail != "" && newEmail != "" {
		err := a.UpdateEmail(w, oldEmail, newEmail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if newPassword != "" {
		err := a.UpdatePassword(w, r, newPassword)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No valid update data provided", http.StatusBadRequest)
		return
	}
	// create kafka message
	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "update_data",
			"user_id": "%d",
			"old_username": "%s",
			"new_username": "%s",
			"new_password": "%s"
			"old_email": "%s",
			"new_email": "%s",
			"timestamp": "%s"
		}`, user.ID, oldUsername, newUsername, newPassword, oldEmail, newEmail, time.Now().UTC().Format(time.RFC3339))),
	}

	// The producer writes the Kafka message to the Kafka cluster
	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		fmt.Println("Failed to produce Kafka message:", err)
	}
}

func (a *App) authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "OAuth authentication failed", http.StatusInternalServerError)
		return
	}

	userExist, err := a.repo.UserExist(a.ctx, nil, user.FirstName, user.Email)
	if err != nil {
		http.Error(w, "Error checking existing user", http.StatusInternalServerError)
		return
	}

	if userExist {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	newUser := repository.User{
		Username: user.FirstName,
		Email:    user.Email,
		Password: "",
	}

	err = newUser.Add(a.ctx, nil)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	message := kafka.Message{
		Value: []byte(fmt.Sprintf(`{
			"event": "signup",
			"user_id": "%s",
			"email": "%s",
			"timestamp": "%s"
		}`, user.UserID, user.Email, time.Now().UTC().Format(time.RFC3339))),
	}

	if err := kafka.ProduceMessage(kafka.Brokers, kafka.Topic, string(message.Value)); err != nil {
		log.Println("Failed to produce Kafka message:", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (a *App) authCallbackHandlerForRouter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a.authCallbackHandler(w, r)
}

func (a *App) InitAuthProviders(r *mux.Router) {
	goth.UseProviders(
		yandex.New(os.Getenv("YANDEX_CLIENT_KEY"), os.Getenv("YANDEX_SECRET"), "http://localhost:4444/auth/yandex/callback"),
		// vk.New("client-id", "client-secret", "http://localhost:4444/auth/vk/callback"),
		github.New(os.Getenv("GITHUB_CLIENT_KEY"), os.Getenv("GITHUB_SECRET"), "http://localhost:4444/auth/github/callback"),
	)

	r.HandleFunc("/auth/{provider}/callback", a.authCallbackHandler)
	r.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler)
}
