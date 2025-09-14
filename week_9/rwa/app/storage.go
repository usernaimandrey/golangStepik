package app

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/argon2"
	"reflect"
	"slices"
	"sort"
	"strings"
)

type Storege struct {
	Users         map[string]*User
	Sessions      map[string]*Session
	NextArticleID int
}

func NewStoreage() *Storege {
	return &Storege{
		Users:         map[string]*User{},
		Sessions:      map[string]*Session{},
		NextArticleID: 1,
	}
}

type User struct {
	Email         string    `json:"email"`
	Password      string    `json:"password"`
	UserName      string    `json:"username"`
	CreatedAt     string    `json:"createdAt"`
	UpdatedAt     string    `json:"updatedAt"`
	EncriptedPass []byte    `json:"-"`
	Token         string    `json:"token"`
	Bio           string    `json:"bio"`
	Articles      []Article `json:"-"`
}

type Session struct {
	Token string `json:"token"`
}

func NewSession(sessioID string) *Session {
	return &Session{Token: sessioID}
}

type UserData struct {
	User User `json:"user"`
}

func NewUserData() *UserData {
	return &UserData{
		User: User{},
	}
}

type Author struct {
	Bio      string `json:"bio"`
	Username string `json:"username"`
}

type Article struct {
	ID             int      `json:"-"`
	Author         Author   `json:"author"`
	Body           string   `json:"body"`
	CreatedAt      string   `json:"createdAt"`
	Description    string   `json:"description"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	Slug           string   `json:"slug" testdiff:"ignore"`
	Tag            []string `json:"tagList"`
	Title          string   `json:"title"`
	UpdatedAt      string   `json:"updatedAt"`
	userId         string   `json:"-"`
}

type ArticlesResponse struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}

func NewArticlesResponse(articles []Article) *ArticlesResponse {
	return &ArticlesResponse{
		Articles:      articles,
		ArticlesCount: len(articles),
	}
}

type ArticleData struct {
	Article Article `json:"article"`
}

func NewArticleData() *ArticleData {
	return &ArticleData{
		Article: Article{},
	}
}

func (st *Storege) CreateUser(user *User) (*User, error) {
	_, err := st.Users[user.Email]

	if err {
		return user, fmt.Errorf("User with email: %s already exists", user.Email)
	}

	if len(user.Email) == 0 || len(user.Password) == 0 {
		return user, fmt.Errorf("password and email can not be blank")
	}

	salt := RandStringRunes(8)

	user.EncriptedPass = user.hashPass(user.Password, salt)
	user.Password = ""
	user.CreatedAt = TimeNowRFC339()
	user.UpdatedAt = TimeNowRFC339()

	st.Users[user.Email] = user
	return user, nil
}

func (st *Storege) CreateArticle(article *Article, user *User) *Article {
	article.ID = st.NextArticleID
	st.NextArticleID += 1
	article.CreatedAt = TimeNowRFC339()
	article.UpdatedAt = TimeNowRFC339()

	article.Slug = strings.Join(strings.Split(strings.ToLower(article.Title), " "), "-")
	article.Author = Author{Bio: user.Bio, Username: user.UserName}

	user.Articles = append(user.Articles, *article)

	return article
}

func (st *Storege) CreateSession(user *User) (*User, error) {

	sessionID := RandStringRunes(32)
	session := NewSession(sessionID)
	st.Sessions[sessionID] = session
	user.Token = sessionID

	return user, nil
}

func (st *Storege) DestroySession(session *Session) error {
	delete(st.Sessions, session.Token)
	user, err := st.UserFindByToken(session.Token)

	if err != nil {
		return err
	}

	user.Token = ""

	return nil
}

func (st *Storege) UserFindByEmail(email string) (*User, error) {
	fakeUser := &User{}
	u, ok := st.Users[email]

	if !ok {
		return fakeUser, fmt.Errorf("user with email %s not found", email)
	}

	return u, nil
}

func (st *Storege) UserFindByToken(token string) (*User, error) {
	var u *User

	for _, user := range st.Users {
		if user.Token == token {
			u = user
		}
	}

	if u == nil {
		return u, fmt.Errorf("user not found")
	}

	return u, nil
}

func (u *User) Update(userData *User) {
	valueData := reflect.ValueOf(userData)
	valueUser := reflect.ValueOf(u).Elem()
	valueDataT := valueData.Elem().Type()
	for i := 0; i < valueDataT.NumField(); i++ {
		field := valueDataT.Field(i)
		nameField := field.Name

		if nameField == "EncriptedPass" || nameField == "Articles" {
			continue
		}

		nameFieldR := valueData.Elem().FieldByName(nameField)
		val := nameFieldR.Interface().(string)

		if len(val) == 0 {
			continue
		}

		nameFieldUserSetable := valueUser.FieldByName(nameField)

		if nameFieldUserSetable.CanSet() {
			nameFieldUserSetable.SetString(val)
		}
	}

	if len(u.UpdatedAt) != 0 {
		u.UpdatedAt = TimeNowRFC339()
	}
}

func (st *Storege) ArticleWhere(params map[string][]string) []Article {
	allArticles := []Article{}

	for _, user := range st.Users {
		allArticles = append(allArticles, user.Articles...)
	}

	if len(params) == 0 {
		sort.Slice(allArticles, func(i, j int) bool { return allArticles[i].ID < allArticles[j].ID })
		return allArticles
	}

	filteredArticles := []Article{}

	for _, article := range allArticles {
		articleVal := reflect.ValueOf(article)
		for param, filter := range params {
			param = capitalizeFirst(param)
			field := articleVal.FieldByName(param)

			if field.IsValid() && field.Kind() == reflect.Struct {
				author := field.Interface().(Author)
				if slices.Contains(filter, author.Bio) || slices.Contains(filter, author.Username) {
					filteredArticles = append(filteredArticles, article)
				}
			} else if field.IsValid() && field.Kind() == reflect.Slice {
				tagList := field.Interface().([]string)
				tagIsIncluded := false
				for _, tag := range filter {
					if slices.Contains(tagList, tag) {
						tagIsIncluded = true
					}
				}
				if tagIsIncluded {
					filteredArticles = append(filteredArticles, article)
				}
			}
		}
	}

	sort.Slice(filteredArticles, func(i, j int) bool { return filteredArticles[i].ID < filteredArticles[j].ID })
	return filteredArticles
}

func (u *User) hashPass(plainPassword, salt string) []byte {
	if len(plainPassword) == 0 {
		return []byte{}
	}
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	res := make([]byte, len(salt))
	copy(res, salt[:len(salt)])
	return append(res, hashedPass...)
}

func (u *User) CheckPass(plainPass string) error {
	salt := string(u.EncriptedPass[0:8])

	hashedPass := u.hashPass(plainPass, salt)

	if !bytes.Equal(hashedPass, u.EncriptedPass) {
		return fmt.Errorf("incorrect login or password")
	}
	return nil
}
