package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"rwa/app"
)

// сюда писать код

type routes map[string]func(w http.ResponseWriter, r *http.Request)

func GetApp() http.Handler {
	router := mux.NewRouter()
	st := app.NewStoreage()
	app := &app.App{
		Router:  router,
		Storege: st,
	}

	app.Router.HandleFunc("/api/users", app.CreateUser).Methods("POST")
	app.Router.HandleFunc("/api/users/login", app.Login).Methods("POST")
	app.Router.HandleFunc("/api/articles", app.Articles).Methods("GET")

	api := app.Router.PathPrefix("/api").Subrouter()
	api.Use(app.AuthMiddleware)
	api.HandleFunc("/user", app.Profile).Methods("GET")
	api.HandleFunc("/user", app.UpdateUser).Methods("PUT")
	api.HandleFunc("/articles", app.ArticleCreate).Methods("POST")
	api.HandleFunc("/user/logout", app.Logout).Methods("POST")
	return app
}
