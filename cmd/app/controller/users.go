package controller

import (
	"AuthDB/cmd/app/repository"
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	users, err := repository.GetAllUsers(ctx, nil)
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
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "Not all fields are filled in", http.StatusBadRequest)
		return
	}
	user, err := repository.NewUser(username, email, password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}
	err = user.Add(ctx, nil)
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
	userID, err := strconv.Atoi(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = user.DeleteByID(ctx, nil, userID)
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
