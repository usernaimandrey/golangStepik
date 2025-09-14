package app

import (
	// "fmt"
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type App struct {
	Router  *mux.Router
	Storege *Storege
}

func (a *App) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Token"))

		session, ok := a.Storege.Sessions[sessionID]

		if !ok {
			http.Error(w, "Not Authtorize", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "Session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Router.ServeHTTP(w, r)
}
