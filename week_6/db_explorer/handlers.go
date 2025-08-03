package main

import (
	"fmt"
	"net/http"
	"sort"
)

func (srv *DbExplorer) RootHandler(w http.ResponseWriter, r *http.Request) {
	entities := []string{}

	for k := range srv.Entities {
		entities = append(entities, k)
	}
	sort.Strings(entities)
	data := map[string]interface{}{
		"response": map[string]interface{}{
			"tables": entities,
		},
	}
	writeJSON(w, http.StatusOK, data)
}

func (srv *DbExplorer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	table, _, err := srv.RouteParse(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
	}

	limitOffset := BuildQueryParams(r.URL.Query())
	sql := fmt.Sprintf("SELECT * FROM %s %s", table, limitOffset)

	rows, err := srv.DataBase.Query(sql)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	defer rows.Close()

	fields := srv.Entities[table].Fields

	responseBuilder := NewResoponseBuilder()
	response, err := responseBuilder.BuildIndexResponse(rows, fields)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (srv *DbExplorer) ShowHandler(w http.ResponseWriter, r *http.Request) {
	table, id, err := srv.RouteParse(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
	}

	field := srv.FetchField(table)

	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = %s", table, field, id)

	rows, err := srv.DataBase.Query(sql)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	fields := srv.Entities[table].Fields
	responseBuilder := NewResoponseBuilder()
	response, rowsCount, err := responseBuilder.BuildShowResponse(rows, fields)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	if rowsCount == 0 {
		RenderError(w, http.StatusNotFound, NewRecordNotFoundError())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (srv *DbExplorer) CreateHandler(w http.ResponseWriter, r *http.Request) {
	table, _, err := srv.RouteParse(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
	}

	data, err := srv.ParseBody(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
	}

	result, err := srv.Insert(data, table)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	id, err := result.LastInsertId()

	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
	}

	var primaryKey string

	for f, v := range srv.Entities[table].Fields {
		if v.Key == "PRI" {
			primaryKey = f
		}
	}

	response := NewResoponseBuilder().BuildCreateResponse(id, primaryKey)

	writeJSON(w, http.StatusOK, response)
}

func (srv *DbExplorer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	table, id, err := srv.RouteParse(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	data, err := srv.ParseBody(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	err = srv.ValidateData(data, table)

	if err != nil {
		RenderError(w, http.StatusBadRequest, err)
		return
	}

	result, err := srv.Update(data, table, id)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, err)
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	response := NewResoponseBuilder().BuildUpdateResponse(rowsAffected)

	writeJSON(w, http.StatusOK, response)
}

func (srv *DbExplorer) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	table, id, err := srv.RouteParse(r)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	result, err := srv.Delete(table, id)

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}
	rowsAffected, err := result.RowsAffected()

	if err != nil {
		RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
		return
	}

	response := NewResoponseBuilder().BuildDeleteResponse(rowsAffected)

	writeJSON(w, http.StatusOK, response)
}
