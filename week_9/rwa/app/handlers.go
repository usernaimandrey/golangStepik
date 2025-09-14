package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	InternalServerError   error = errors.New("internal error")
	UserUnauthtorizeError error = errors.New("User not authtorize")
)

func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	userData, err := a.UserRequestParse(r)

	if err != nil {
		RenderError(w, err, http.StatusInternalServerError)
		return
	}

	user, err := a.Storege.CreateUser(userData)

	if err != nil {
		RenderError(w, err, http.StatusConflict)
		return
	}

	user, err = a.Storege.CreateSession(user)

	if err != nil {
		RenderError(w, err, http.StatusConflict)
		return
	}

	RenderResponse(w, "user", user, http.StatusCreated)
}

func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value("Session").(*Session)

	if !ok {
		RenderError(w, UserUnauthtorizeError, http.StatusUnauthorized)
		return
	}

	user, err := a.Storege.UserFindByToken(session.Token)

	if err != nil {
		RenderError(w, err, http.StatusNotFound)
		return
	}

	userData, err := a.UserRequestParse(r)

	if err != nil {
		RenderError(w, err, http.StatusNotFound)
		return
	}

	user.Update(userData)

	RenderResponse(w, "user", user, http.StatusOK)
}

func (a *App) UserRequestParse(r *http.Request) (*User, error) {
	userData := NewUserData()
	data, err := io.ReadAll(r.Body)

	if err != nil {
		return &userData.User, InternalServerError
	}

	json.Unmarshal(data, userData)
	return &userData.User, nil
}

func (a *App) ArticleRequestParse(r *http.Request) (*Article, error) {
	articleData := NewArticleData()
	data, err := io.ReadAll(r.Body)

	if err != nil {
		return &articleData.Article, err
	}

	json.Unmarshal(data, articleData)
	return &articleData.Article, nil
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	userData, err := a.UserRequestParse(r)

	if err != nil {
		RenderError(w, err, http.StatusInternalServerError)
		return
	}

	user, err := a.Storege.UserFindByEmail(userData.Email)

	if err != nil {
		RenderError(w, err, http.StatusNotFound)
		return
	}

	err = user.CheckPass(userData.Password)

	if err != nil {
		RenderError(w, err, http.StatusBadRequest)
		return
	}

	user, err = a.Storege.CreateSession(user)

	if err != nil {
		RenderError(w, InternalServerError, http.StatusBadRequest)
		return
	}

	RenderResponse(w, "user", user, http.StatusOK)
}

func (a *App) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value("Session").(*Session)

	if !ok {
		RenderError(w, UserUnauthtorizeError, http.StatusUnauthorized)
		return
	}

	user, err := a.Storege.UserFindByToken(session.Token)

	if err != nil {
		RenderError(w, err, http.StatusNotFound)
		return
	}

	RenderResponse(w, "user", user, http.StatusOK)
}

func (a *App) ArticleCreate(w http.ResponseWriter, r *http.Request) {
	articleData, err := a.ArticleRequestParse(r)

	if err != nil {
		RenderError(w, err, http.StatusInternalServerError)
	}

	ctx := r.Context()
	session, ok := ctx.Value("Session").(*Session)

	if !ok {
		RenderError(w, UserUnauthtorizeError, http.StatusUnauthorized)
		return
	}

	user, err := a.Storege.UserFindByToken(session.Token)

	article := a.Storege.CreateArticle(articleData, user)

	RenderResponse(w, "article", article, http.StatusCreated)
}

func (a *App) Articles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := map[string][]string{}

	for paramName, values := range query {
		if _, ok := params[paramName]; !ok {
			params[paramName] = values
		} else {
			params[paramName] = append(params[paramName], values...)
		}
	}

	articles := a.Storege.ArticleWhere(params)

	responseData := NewArticlesResponse(articles)
	RenderResponse(w, "", responseData, http.StatusOK)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := ctx.Value("Session").(*Session)

	if !ok {
		RenderError(w, UserUnauthtorizeError, http.StatusUnauthorized)
		return
	}

	err := a.Storege.DestroySession(session)

	if err != nil {
		RenderError(w, err, http.StatusNotFound)
	}

	w.WriteHeader(http.StatusOK)
}
