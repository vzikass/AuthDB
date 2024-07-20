package controller

import (
	"context"
	"encoding/json"
	"exercise/app/repository"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	users, err := repository.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	main := filepath.Join("public", "html", "usersPage.html")
	tmpl, err := template.ParseFiles(main)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	err = tmpl.ExecuteTemplate(w, "users", users)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}

func AddUsers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	login := r.FormValue("login")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if login == "" || email == "" || password == "" {
		http.Error(w, "not all fields are filled in", http.StatusBadRequest)
		return
	}
	user, err := repository.NewUser(login, email, password)
	if err != nil{
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}
	err = user.Add(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = json.NewEncoder(w).Encode("User added successfully")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func DeleteUserByID(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	userId := p.ByName("userID")
	ctx := context.Background()
	user, err := repository.GetUserById(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = user.Delete(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode("User deleted successfully")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func UpdateUserByID(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	userId := p.ByName("userID")
	login := r.FormValue("login")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := repository.GetUserById(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.Login = login
	user.Email = email
	user.Password = password
	err = user.Update(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode("Data updated")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}