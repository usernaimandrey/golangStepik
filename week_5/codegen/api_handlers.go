package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)


func RenderError(w http.ResponseWriter, status int, err error) {
	data := map[string]interface{}{
		"error": "",
	}

	switch err.(type) {
	case ApiError:
		errCast := err.(ApiError)
		data["error"] = errCast.Error()
		status = errCast.HTTPStatus
	default:
		data["error"] = err.Error()
	}

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
	}
	w.WriteHeader(status)
	w.Write(resp)
}


// MyApiProfileWraper
func (srv *MyApi) MyApiProfileWraper(w http.ResponseWriter, r *http.Request) {


  params := &ProfileParams{}

  login := r.FormValue("login")
	
	
	
  params.Login = login
	
  
	
		
  if len(params.Login) == 0 {
	  err := fmt.Errorf("login must me not empty")
	  RenderError(w, http.StatusBadRequest, err)
	  return
  }
		
		
		
	

  user, err := srv.Profile(r.Context(), *params)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]interface{}{
		"error":    "",
		"response": user,
	}

	data, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{error:" + err.Error() + "}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// MyApiCreateWraper
func (srv *MyApi) MyApiCreateWraper(w http.ResponseWriter, r *http.Request) {

	method := r.Method
	if method != http.MethodPost {
		RenderError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))
		return
	}


  if r.Header.Get("X-Auth") != "100500" {
		RenderError(w, http.StatusForbidden, fmt.Errorf("unauthorized"))
		return
	}

  params := &CreateParams{}

  login := r.FormValue("login")
	
	
	
  params.Login = login
	
  
	
		
  if len(params.Login) == 0 {
	  err := fmt.Errorf("login must me not empty")
	  RenderError(w, http.StatusBadRequest, err)
	  return
  }
		
		
			
	if len(params.Login) < 10 {
		err := fmt.Errorf("login len must be >= 10")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
		
	

  full_name := r.FormValue("full_name")
	
	
	
  params.Name = full_name
	
  
	
		
		
		
	

  status := r.FormValue("status")
	
	
	
  params.Status = status
	
  
	if params.Status == "" {
		params.Status = "user"
	}
	validParam := false
	for _, v := range []string{ "user",  "moderator",  "admin"  } {
		if v == params.Status {
			validParam = true
		}
	}
	if !validParam {
		err := fmt.Errorf("status must be one of [user, moderator, admin]")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
	
	

  age := r.FormValue("age")
	
	
	
  var ageInt int
	if len(age) != 0 {
		var err error
		ageInt, err = strconv.Atoi(r.FormValue("age"))
		if err != nil {
			RenderError(w, http.StatusBadRequest, fmt.Errorf("age must be int"))
			return
		}
	}
  params.Age = ageInt

	
  
	
		
		
			
	if params.Age < 0 {
		err := fmt.Errorf("age must be >= 0")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
		
		  
	if params.Age > 128 {
		err := fmt.Errorf("age must be <= 128")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
	

  user, err := srv.Create(r.Context(), *params)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]interface{}{
		"error":    "",
		"response": user,
	}

	data, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{error:" + err.Error() + "}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// MyApi
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	
	case "/user/profile":
		srv.MyApiProfileWraper(w, r)
	
	case "/user/create":
		srv.MyApiCreateWraper(w, r)
	
	default:
		RenderError(w, http.StatusNotFound, fmt.Errorf("unknown method"))
	}
}

// OtherApiCreateWraper
func (srv *OtherApi) OtherApiCreateWraper(w http.ResponseWriter, r *http.Request) {

	method := r.Method
	if method != http.MethodPost {
		RenderError(w, http.StatusNotAcceptable, fmt.Errorf("bad method"))
		return
	}


  if r.Header.Get("X-Auth") != "100500" {
		RenderError(w, http.StatusForbidden, fmt.Errorf("unauthorized"))
		return
	}

  params := &OtherCreateParams{}

  username := r.FormValue("username")
	
	
	
  params.Username = username
	
  
	
		
  if len(params.Username) == 0 {
	  err := fmt.Errorf("username must me not empty")
	  RenderError(w, http.StatusBadRequest, err)
	  return
  }
		
		
			
	if len(params.Username) < 3 {
		err := fmt.Errorf("username len must be >= 3")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
		
	

  account_name := r.FormValue("account_name")
	
	
	
  params.Name = account_name
	
  
	
		
		
		
	

  class := r.FormValue("class")
	
	
	
  params.Class = class
	
  
	if params.Class == "" {
		params.Class = "warrior"
	}
	validParam := false
	for _, v := range []string{ "warrior",  "sorcerer",  "rouge"  } {
		if v == params.Class {
			validParam = true
		}
	}
	if !validParam {
		err := fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
	
	

  level := r.FormValue("level")
	
	
	
  var levelInt int
	if len(level) != 0 {
		var err error
		levelInt, err = strconv.Atoi(r.FormValue("level"))
		if err != nil {
			RenderError(w, http.StatusBadRequest, fmt.Errorf("level must be int"))
			return
		}
	}
  params.Level = levelInt

	
  
	
		
		
			
	if params.Level < 1 {
		err := fmt.Errorf("level must be >= 1")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
		
		  
	if params.Level > 50 {
		err := fmt.Errorf("level must be <= 50")
		RenderError(w, http.StatusBadRequest, err)
		return
	}
			
		
	

  user, err := srv.Create(r.Context(), *params)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]interface{}{
		"error":    "",
		"response": user,
	}

	data, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{error:" + err.Error() + "}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// OtherApi
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	
	case "/user/create":
		srv.OtherApiCreateWraper(w, r)
	
	default:
		RenderError(w, http.StatusNotFound, fmt.Errorf("unknown method"))
	}
}
