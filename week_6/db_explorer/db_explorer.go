package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

import (
	"database/sql"
	"fmt"
	"net/http"
)

const (
	defaultLimit        string = "5"
	defaultOffset       string = "0"
	routeIDPttern       string = `\d+`
	routeParsingPattern string = `^/([^/]+)/([^/]+)$`
)

type DbExplorer struct {
	DataBase *sql.DB
	Router   map[string]*Route
	Entities map[string]*Table
}

type Table struct {
	Name   string
	Fields map[string]*Field
}

type Field struct {
	Name       string
	Type       string
	Collation  interface{}
	Null       string
	Key        string
	Default    interface{}
	Extra      string
	Privileges string
	Comment    string
}

type Handler func(w http.ResponseWriter, r *http.Request)

type Route struct {
	Handler Handler
}

func (srv *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	route, err := srv.FetchRoute(r)

	if err != nil {
		RenderError(w, http.StatusNotFound, err)
		return
	}
	route.Handler(w, r)
}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	ex := &DbExplorer{DataBase: db, Entities: map[string]*Table{}}
	rows, err := ex.DataBase.Query("SHOW TABLES")

	if err != nil {
		return ex, err
	}

	var scanError error

	defer rows.Close()

	for rows.Next() {
		table := &Table{Fields: map[string]*Field{}}
		err := rows.Scan(&table.Name)

		if err != nil {
			scanError = err
			break
		}
		ex.Entities[table.Name] = table
	}

	if scanError != nil {
		return ex, scanError
	}

	for _, table := range ex.Entities {
		err := ex.NewField(table)

		if err != nil {
			scanError = err
			break
		}
	}

	if scanError != nil {
		return ex, scanError
	}

	ex.SetRoutes()

	return ex, nil
}

func (srv *DbExplorer) SetRoutes() {
	routes := map[string]*Route{
		"/-GET": &Route{Handler: srv.RootHandler},
	}

	for k := range srv.Entities {
		indexRoute := fmt.Sprintf("/%s-GET", k)
		showRoute := fmt.Sprintf("/%s/id-GET", k)
		createRoute := fmt.Sprintf("/%s/-PUT", k)
		updateRoute := fmt.Sprintf("/%s/id-POST", k)
		deleteRoute := fmt.Sprintf("/%s/id-DELETE", k)

		routes[indexRoute] = &Route{Handler: srv.IndexHandler}
		routes[showRoute] = &Route{Handler: srv.ShowHandler}
		routes[createRoute] = &Route{Handler: srv.CreateHandler}
		routes[updateRoute] = &Route{Handler: srv.UpdateHandler}
		routes[deleteRoute] = &Route{Handler: srv.DeleteHandler}
	}

	srv.Router = routes
}

func (srv *DbExplorer) NewField(table *Table) error {
	sql := fmt.Sprintf("SHOW FULL COLUMNS FROM %s", table.Name)
	rowsTable, err := srv.DataBase.Query(sql)

	if err != nil {
		return err
	}

	defer rowsTable.Close()

	var scanError error

	for rowsTable.Next() {
		field := &Field{}
		err := rowsTable.Scan(
			&field.Name,
			&field.Type,
			&field.Collation,
			&field.Null,
			&field.Key,
			&field.Default,
			&field.Extra,
			&field.Privileges,
			&field.Comment,
		)

		if err != nil {
			scanError = err
			break
		}
		table.Fields[field.Name] = field
	}
	return scanError
}
