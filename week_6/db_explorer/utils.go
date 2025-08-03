package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, statusCode int, data map[string]interface{}) {
	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resp)
}

func IsEmptyDb(en []string) bool {
	return len(en) == 0
}

func RenderError(w http.ResponseWriter, status int, err error) {
	data := map[string]interface{}{
		"error": err.Error(),
	}

	writeJSON(w, status, data)
}

func DataTypeConverter(value string, typeValue string) (interface{}, error) {
	var result interface{}
	var err error

	switch typeValue {
	case "int":
		result, err = strconv.Atoi(value)
	case "float":
		result, err = strconv.ParseFloat(value, 64)
	case "double":
		result, err = strconv.ParseFloat(value, 64)
	default:
		result = value
	}
	return result, err
}
