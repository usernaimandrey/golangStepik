package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	longForm                    = "2006-01-02T15:04:05 -03:00"
	AccessToken                 = "asdfasdf"
	ResponseInternalServerError = `{"status": 500, "Error": "Internal Error"}`
	ResponseUnauthorized        = `{"status": 401, "Error": "Bad AccessToken"}`
	ResponseBadRequest          = `{"status": 400, "Error": "ErrorBadOrderField"}`
	BrockenJson                 = `{"status": 400`
	UnpackJson                  = `{"status": 400}`
)

type DataXml struct {
	XMLName xml.Name      `xml:"root"`
	User    []UserDataRow `xml:"row"`
}

type UserDataRow struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type TestCase struct {
	ID            string
	AccessToken   string
	Url           string
	SearchRequest *SearchRequest
	Result        *SearchResponse
	IsError       bool
}

func (u *UserDataRow) RegisteredToTime() (time.Time, error) {
	t, err := time.Parse(longForm, u.Registered)
	if err != nil {
		return time.Now(), err
	}

	return t, nil
}

func (u *UserDataRow) FullName() string {
	return u.LastName + " " + u.FirstName
}

func authorize(token string) error {
	if token == AccessToken {
		return nil
	}
	return fmt.Errorf("access denied")
}

func orderFieldValidator(orderField string) error {
	orderFields := []string{
		"Id", "Guid", "IsActive", "Balance", "Picture", "Age", "EyeColor", "FirstName", "LastName", "Gender", "Company", "Email", "Phone", "Address", "About", "Registered", "FavoriteFruit",
	}

	isContain := false
	for _, field := range orderFields {
		if field == orderField {
			isContain = true
		}
	}

	if !isContain {
		return fmt.Errorf("not contain")
	}
	return nil
}

func renderError(w http.ResponseWriter, status int, brockenJson bool) {
	payload := ""
	switch status {
	case http.StatusUnauthorized:
		payload = ResponseUnauthorized
	case http.StatusBadRequest:
		payload = ResponseBadRequest
	case http.StatusInternalServerError:
		payload = ResponseInternalServerError
	default:
		payload = ResponseInternalServerError
	}

	if brockenJson {
		payload = BrockenJson
	}

	w.WriteHeader(status)
	io.WriteString(w, payload)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	err := authorize(token)

	if err != nil {
		renderError(w, http.StatusUnauthorized, false)
		return
	}

	query := r.FormValue("query")
	limit := r.FormValue("limit")
	offset := r.FormValue("offset")
	orderBy := r.FormValue("order_by")
	orderField := r.FormValue("order_field")

	if query == "_internal_error" {
		renderError(w, http.StatusInternalServerError, false)
		return
	}
	if query == "_broken_json" {
		renderError(w, http.StatusBadRequest, true)
		return
	}

	err = orderFieldValidator(orderField)

	if err != nil {
		renderError(w, http.StatusBadRequest, false)
		return
	}

	f, err := os.Open("dataset.xml")

	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}

	data, err := io.ReadAll(f)

	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}

	users := new(DataXml)
	err = xml.Unmarshal(data, users)

	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}

	filteredUsers := []User{}

	for _, user := range users.User {
		v := reflect.ValueOf(&user).Elem()
		field := v.FieldByName(orderField)
		if field.IsValid() {
			typeField := field.Type().String()
			if typeField == "int" {
				if strconv.Itoa(int(field.Int())) == query {
					filteredUsers = append(filteredUsers, User{Id: user.Id, Name: user.FullName(), Age: user.Age, About: user.About, Gender: user.Gender})
				}
			} else if typeField == "string" {
				if strings.Contains(field.String(), query) {
					filteredUsers = append(filteredUsers, User{Id: user.Id, Name: user.FullName(), Age: user.Age, About: user.About, Gender: user.Gender})
				}
			}
		}
	}
	limitInt, err := strconv.Atoi(limit)

	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}

	offsetInt, err := strconv.Atoi(offset)

	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}

	if (limitInt + offsetInt) <= len(filteredUsers) {
		filteredUsers = filteredUsers[offsetInt:(limitInt + offsetInt)]
	}

	if orderBy == "1" {
		fmt.Println("sort")
		sort.Slice(filteredUsers, func(i, j int) bool {
			return filteredUsers[i].Id < filteredUsers[j].Id
		})
	} else if orderBy == "-1" {
		sort.Slice(filteredUsers, func(i, j int) bool {
			return filteredUsers[i].Id > filteredUsers[j].Id
		})
	}

	payload, err := json.Marshal(filteredUsers)
	if err != nil {
		renderError(w, http.StatusInternalServerError, false)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(payload))
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	cases := []TestCase{
		TestCase{
			ID:  "_status_unauthorized",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "1",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: "asdf",
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_limit_negative",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "1",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_broken_url",
			Url: "",
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "1",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_broken_address",
			Url: "http://ex.com",
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "1",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_offset_negative",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     -1,
				Query:      "1",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_internal_error",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "_internal_error",
				OrderField: "id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_status_bad_request",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "true",
				OrderField: "isPassive",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_broken_json",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "_broken_json",
				OrderField: "isActive",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			ID:  "_by_id",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "1",
				OrderField: "Id",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result: &SearchResponse{
				Users: []User{User{
					Id:     1,
					Name:   "Mayer Hilda",
					Age:    21,
					About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
					Gender: "female",
				}},
			},
			IsError: false,
		},
		TestCase{
			ID:  "_by_about",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      30,
				Offset:     0,
				Query:      "Adipisicing",
				OrderField: "About",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     3,
						Name:   "Dillard Everett",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
					User{
						Id:     17,
						Name:   "Mccoy Dillard",
						Age:    36,
						About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
						Gender: "male",
					},
					User{
						Id:     20,
						Name:   "York Lowery",
						Age:    27,
						About:  "Dolor enim sit id dolore enim sint nostrud deserunt. Occaecat minim enim veniam proident mollit Lorem irure ex. Adipisicing pariatur adipisicing aliqua amet proident velit. Magna commodo culpa sit id.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			ID:  "_by_about_with_limit",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "Adipisicing",
				OrderField: "About",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     3,
						Name:   "Dillard Everett",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			ID:  "_by_about_with_offset",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     1,
				Query:      "Adipisicing",
				OrderField: "About",
				OrderBy:    0,
			},
			AccessToken: AccessToken,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     17,
						Name:   "Mccoy Dillard",
						Age:    36,
						About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			ID:  "_by_about_with_sort",
			Url: ts.URL,
			SearchRequest: &SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "Adipisicing",
				OrderField: "About",
				OrderBy:    1,
			},
			AccessToken: AccessToken,
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     3,
						Name:   "Dillard Everett",
						Age:    27,
						About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
					User{
						Id:     17,
						Name:   "Mccoy Dillard",
						Age:    36,
						About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
	}

	for caseNum, item := range cases {
		client := SearchClient{
			AccessToken: item.AccessToken,
			URL:         item.Url,
		}

		r, err := client.FindUsers(*item.SearchRequest)

		fmt.Println(r, err)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Result, r) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, r)
		}
	}
	ts.Close()
}
