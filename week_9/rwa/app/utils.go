package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	sizes       = []uint{80, 160, 320}
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RenderError(w http.ResponseWriter, err error, errCode int) {
	w.WriteHeader(errCode)
	errData := map[string]interface{}{
		"error": map[string]interface{}{
			"stsus":   errCode,
			"details": err.Error(),
		},
	}
	data, _ := json.Marshal(errData)
	w.Write(data)
}

func RenderResponse(w http.ResponseWriter, rootKey string, data interface{}, status int) {
	var responseData interface{}

	if len(rootKey) == 0 {
		responseData = data
	} else {
		responseData = map[string]interface{}{rootKey: data}
	}
	response, err := json.Marshal(responseData)

	if err != nil {
		RenderError(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(response)
}

func TimeNowRFC339() string {
	now := time.Now()
	rfc3339Time := now.Format(time.RFC3339)
	return rfc3339Time
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
