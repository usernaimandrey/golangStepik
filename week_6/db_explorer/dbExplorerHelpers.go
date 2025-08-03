package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func (srv *DbExplorer) ParseBody(r *http.Request) (map[string]interface{}, error) {
	var data interface{}
	body, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		return map[string]interface{}{}, err
	}

	json.Unmarshal(body, &data)

	return data.(map[string]interface{}), nil
}

func (srv *DbExplorer) FetchRoute(r *http.Request) (*Route, error) {
	path := r.URL.Path
	method := r.Method
	re, err := regexp.Compile(routeIDPttern)

	if err != nil {
		return &Route{}, err
	}

	path = re.ReplaceAllString(path, "id")
	routeKey := fmt.Sprintf("%s-%s", path, method)
	route, ok := srv.Router[routeKey]

	if !ok {
		return route, fmt.Errorf("unknown table")
	}

	return route, nil
}

func (srv *DbExplorer) FetchField(tableName string) string {
	table := srv.Entities[tableName]
	var field string

	for k, _ := range table.Fields {
		if strings.Contains(k, "id") {
			field = k
		}
	}
	return field
}

func BuildQueryParams(q url.Values) string {
	limit := q.Get("limit")
	offset := q.Get("offset")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l < 0 {
			limit = defaultLimit
		}
	} else {
		limit = defaultLimit
	}

	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil || o < 0 {
			offset = defaultOffset
		}
	} else {
		offset = defaultOffset
	}

	return fmt.Sprintf("LIMIT %s OFFSET %s", limit, offset)
}

func (srv *DbExplorer) RouteParse(r *http.Request) (string, string, error) {
	path := r.URL.Path
	re, err := regexp.Compile(routeParsingPattern)

	if err != nil {
		return "", "", err
	}

	matches := re.FindStringSubmatch(path)
	var id string
	var table string

	if len(matches) != 3 {
		table = strings.ReplaceAll(path, "/", "")
		return table, "", nil
	}

	table = matches[1]
	id = matches[2]

	return table, id, nil
}

func (srv *DbExplorer) ValidateData(data map[string]interface{}, table string) error {
	var err error

	for f, v := range data {
		field := srv.Entities[table].Fields[f]
		if field.Extra == "auto_increment" {
			err = NewFieldInvalidTypeError(field.Name)
			break
		}

		fieldIsNull := field.Null == "YES"

		if v == nil {
			if fieldIsNull {
				continue
			} else {
				err = NewFieldInvalidTypeError(field.Name)
				break
			}
		}

		fieldType := field.Type

		if fieldType == "varchar(255)" || fieldType == "text" {
			_, ok := v.(string)
			if !ok {
				err = NewFieldInvalidTypeError(field.Name)
				break
			}
		} else if fieldType == "int" {
			_, ok := v.(int)
			if !ok {
				err = NewFieldInvalidTypeError(field.Name)
				break
			}
		} else if fieldType == "float" || fieldType == "double" {
			_, ok := v.(float64)
			if !ok {
				err = NewFieldInvalidTypeError(field.Name)
				break
			}
		}
	}
	return err
}

func (srv *DbExplorer) Insert(data map[string]interface{}, table string) (sql.Result, error) {
	fields := []string{}
	valuesWithPlaceHolder := "VALUES("
	values := []interface{}{}
	nullTemplate := `%s %s, `

	for field, prop := range srv.Entities[table].Fields {
		v, ok := data[field]
		if !ok {
			if prop.Null == "YES" {
				continue
			} else if prop.Type == "text" || prop.Type == "varchar(255)" {
				v = ""
			} else {
				v = 0
			}
		}

		if prop.Extra == "auto_increment" {
			continue
		}

		if v != nil {
			values = append(values, v)
		} else {
			values = append(values, "NULL")
		}

		valuesWithPlaceHolder = fmt.Sprintf(nullTemplate, valuesWithPlaceHolder, "?")

		fields = append(fields, prop.Name)
	}
	sql := fmt.Sprintf(`INSERT INTO %s (%s) %s)`, table, strings.Join(fields, ", "), strings.TrimRight(valuesWithPlaceHolder, ", "))

	return srv.DataBase.Exec(sql, values...)
}

func (srv *DbExplorer) Update(data map[string]interface{}, table string, id string) (sql.Result, error) {
	fields := []string{}
	values := []interface{}{}
	template := `%s = ?`

	for extField, v := range data {
		fieldAndValue := ""
		field, ok := srv.Entities[table].Fields[extField]

		if !ok {
			continue
		}

		if v != nil {
			values = append(values, v)
		} else {
			values = append(values, nil)
		}

		fieldAndValue = fmt.Sprintf(template, field.Name)

		fields = append(fields, fieldAndValue)
	}

	fieldFromCondition := srv.FetchField(table)

	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE %s = %s`, table, strings.Join(fields, ", "), fieldFromCondition, id)

	return srv.DataBase.Exec(sql, values...)
}

func (srv *DbExplorer) Delete(table string, id string) (sql.Result, error) {
	sql := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table)

	return srv.DataBase.Exec(sql, id)
}
