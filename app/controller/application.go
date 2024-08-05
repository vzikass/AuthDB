package controller

import (
	"context"
	"AuthDB/app/repository"
	"AuthDB/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
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
		a.renderDeleteConfirmationPage(w)
	}))
	r.POST("/update", a.authorized(a.UpdateData))
	r.GET("/update", a.authorized(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdateUserPage(w, "")
	}))
	r.GET("/logout", a.authorized(a.Logout))

	r.GET("/users", a.authorized(GetAllUsers))
}

func (a *App) LoginPage(w http.ResponseWriter, message string) {
	path := filepath.Join("public", "html", "login.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	type answer struct {
		Message string
	}
	data := answer{Message: message}
	err = tmpl.ExecuteTemplate(w, "login", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *App) Login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := r.FormValue("login")
	password := r.FormValue("password")
	if login == "" || password == "" {
		a.LoginPage(w, "You must provide a login and password")
		return
	}
	user, err := a.repo.Login(a.ctx, login)
	if err != nil {
		a.LoginPage(w, "User not found")
		return
	}
	if !utils.CompareHashPassword(password, user.Password) {
		a.LoginPage(w, "Incorrect password")
		return
	}
	token, err := utils.GenerateToken(user.Login)
	if err != nil{
		log.Fatalf("Error generate token: %v", err)
		return
	}
	hashedToken, err := utils.GenerateHash(token)
	if err != nil {
		log.Fatalf("Error generate hash token: %v\n", err)
		return
	}
	// to protect access to hash
	a.cacheMu.Lock()
	a.cache[hashedToken] = user
	a.cacheMu.Unlock()
	// Create cookies when login
	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)
	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(hashedToken),
		Expires: expiration, Secure: true, HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) SignupPage(w http.ResponseWriter, message string) {
	path := filepath.Join("public", "html", "signup.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	type answer struct {
		Message string
	}
	data := answer{Message: message}
	err = tmpl.ExecuteTemplate(w, "signup", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func isNumeric(s string) bool {
	// numbers from 0 to 9 matches the previous token between one and unlimited times
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

func (a *App) Signup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := strings.TrimSpace(r.FormValue("login"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))
	if login == "" || email == "" || password == "" || password2 == "" {
		a.SignupPage(w, "Not all fields are filled in")
		return
	}
	if isNumeric(login) {
		a.SignupPage(w, "Login cannot be entirely numeric")
		return
	}
	if password != password2 {
		a.SignupPage(w, "Password mismatch")
		return
	}
	if len(password) <= 3 || len(login) <= 3 {
		a.SignupPage(w, "Minimum field length - 4 characters")
		return
	}
	userExist, err := a.repo.UserExist(a.ctx, login, email)
	if err != nil {
		a.SignupPage(w, "Error checking existing user")
		return
	}
	if userExist {
		a.SignupPage(w, "User already created")
		return
	}
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		user, err := repository.NewUser(login, email, password)
		if err != nil {
			errCh <- err
			return
		}
		err = user.Add(a.ctx)
		if err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	err = <-errCh
	if err != nil {
		a.SignupPage(w, err.Error())
		return
	}
	a.LoginPage(w, fmt.Sprintln("Successful signup!"))
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// To exit we can delete cookies
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func ReadCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		return value, err
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return value, err
	}
	str := cookie.Value
	value, err = url.QueryUnescape(str)
	if err != nil {
		return value, err
	}
	return value, err
}

func (a *App) authorized(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		token, err := ReadCookie("token", r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		_, ok := a.cache[token]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r, p)
	}
}

func (a *App) HomePage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	path := filepath.Join("public", "html", "index.html")
	path2 := filepath.Join("public", "html", "delete.html")
	path3 := filepath.Join("public", "html", "update.html")
	tmpl, err := template.ParseFiles(path, path2, path3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *App) renderDeleteConfirmationPage(w http.ResponseWriter) {
	path := filepath.Join("public", "html", "delete.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}
}

func (a *App) DeleteAccount(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	token, err := ReadCookie("token", r)
	if err != nil {
		log.Printf("Error reading token cookie: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, ok := a.cache[token]
	if !ok {
		log.Printf("Token not found in cache")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = a.repo.DeleteUserByID(a.ctx, user.ID)
	if err != nil {
		log.Printf("Error deleting user by ID: %v", err)
		http.Error(w, "Something went wrong, please try later", http.StatusInternalServerError)
		return
	}
	delete(a.cache, token)

	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) UpdateUserPage(w http.ResponseWriter, message string) {
	path := filepath.Join("public", "html", "update.html")
	path2 := filepath.Join("public", "html", "login.html")
	tmpl, err := template.ParseFiles(path, path2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (a *App) UpdateLogin(w http.ResponseWriter, oldLogin, newLogin string) error {
	user, err := a.repo.FindUserByLogin(a.ctx, oldLogin)
	if err != nil {
		a.UpdateUserPage(w, "User not found")
		return err
	}
	if user.Login != newLogin {
		query := `UPDATE users SET login = $1 WHERE login = $2`
		err = a.repo.UpdateData(a.ctx, query, newLogin, oldLogin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil
	}
	a.UpdateUserPage(w, "This login already exists")
	return fmt.Errorf("login already exists")
}

func (a *App) UpdateEmail(w http.ResponseWriter, oldEmail, newEmail string) error {
	user, err := a.repo.FindUserByEmail(a.ctx, oldEmail)
	if err != nil {
		a.UpdateUserPage(w, "User not found")
		return err
	}
	if user.Email != newEmail {
		query := `UPDATE users SET email = $1 WHERE email = $2`
		err = a.repo.UpdateData(a.ctx, query, newEmail, oldEmail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil
	}
	a.UpdateUserPage(w, "This email already exists")
	return fmt.Errorf("email already exists")
}

func (a *App) UpdatePassword(w http.ResponseWriter, r *http.Request, newPassword string) error {
	userHashedPass := repository.HashPassword
	if len(newPassword) <= 3 {
		a.UpdateUserPage(w, "Minimum field length - 4 characters")
	}
	user, err := a.repo.FindUserByPassword(a.ctx, userHashedPass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	hashedNewPassword, err := utils.GenerateHash(newPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if user.Password != newPassword{
		query := `UPDATE users SET password = $1 WHERE password = $2`
		err = a.repo.UpdateData(a.ctx, query, hashedNewPassword, userHashedPass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	user.Password = hashedNewPassword
	return err
}

func (a *App) UpdateData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	oldLogin := r.FormValue("oldLogin")
	newLogin := r.FormValue("newLogin")
	oldEmail := r.FormValue("oldEmail")
	newEmail := r.FormValue("newEmail")
	newPassword := r.FormValue("newPassword")

	if oldLogin != "" && newLogin != "" {
		err := a.UpdateLogin(w, oldLogin, newLogin)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
}
