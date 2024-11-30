// This file contains the rendering of html templates
package controller

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

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

func (a *App) RenderDeleteConfirmationPage(w http.ResponseWriter) {
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

func (a *App) HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		filepath.Join("public", "html", "main.html"),
		filepath.Join("public", "html", "delete.html"),
		filepath.Join("public", "html", "update.html"),
	)
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
