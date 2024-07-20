package controller

import (
	"context"
	"exercise/app/repository"
	"exercise/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
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
	r.GET("/", a.authorized(a.HomePage))

	r.GET("/logout", a.authorized(a.Logout))

	// this is working with a database, so authorization is required
	r.GET("/users", a.authorized(GetAllUsers))
	r.POST("/users/add", a.authorized(AddUsers))
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
	token := utils.GenerateRandomToken()
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

func (a *App) Signup(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := strings.TrimSpace(r.FormValue("login"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))
	if login == "" || email == "" || password == "" || password2 == "" {
		a.SignupPage(w, "Not all fields are filled in")
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
	value, _ = url.QueryUnescape(str)
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
	tmpl, err := template.ParseFiles(path)
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

func (a *App) renderDeleteConfirmationPage(w http.ResponseWriter){
	path := filepath.Join("public", "html", "delete.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}
}

func (a *App) DeleteAccount(w http.ResponseWriter, r *http.Request, p httprouter.Params){
	token, err := ReadCookie("token", r)
	if err != nil{
		log.Printf("Error reading token cookie: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, ok := a.cache[token]
	if !ok{
		log.Printf("Token not found in cache")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	err = a.repo.DeleteUserByID(a.ctx, user.ID)
	if err != nil{
		log.Printf("Error deleting user by ID: %v", err)
		http.Error(w, "Something went wrong, please try later", http.StatusInternalServerError)
		return
	}
	delete(a.cache, token)
	for _, v := range r.Cookies(){
		c := http.Cookie{
			Name: v.Name,
			MaxAge: -1,
		}
		http.SetCookie(w, &c)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}