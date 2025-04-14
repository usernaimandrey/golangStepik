package main

// import (
// 	"encoding/xml"
// 	"fmt"
// 	"io"
// 	"os"
// 	"time"
// )

// const longForm = "2006-01-02T15:04:05 -03:00"

// type Users struct {
// 	XMLName xml.Name `xml:"root"`
// 	User    []User   `xml:"row"`
// }

// type User struct {
// 	Id            int    `xml:"id"`
// 	Guid          string `xml:"guid"`
// 	IsActive      bool   `xml:"isActive"`
// 	Balance       string `xml:"balance"`
// 	Picture       string `xml:"picture"`
// 	Age           int    `xml:"age"`
// 	EyeColor      string `xml:"eyeColor"`
// 	FirstName     string `xml:"first_name"`
// 	LastName      string `xml:"last_name"`
// 	Gender        string `xml:"gender"`
// 	Company       string `xml:"company"`
// 	Email         string `xml:"email"`
// 	Phone         string `xml:"phone"`
// 	Address       string `xml:"address"`
// 	About         string `xml:"about"`
// 	Registered    string `xml:"registered"`
// 	FavoriteFruit string `xml:"favoriteFruit"`
// }

// func (u *User) RegisteredToTime() (time.Time, error) {
// 	t, err := time.Parse(longForm, u.Registered)
// 	if err != nil {
// 		return time.Now(), err
// 	}

// 	return t, nil
// }

// func (u *User) FullName() string {
// 	return u.LastName + " " + u.FirstName
// }

func main() {
	// f, err := os.Open("dataset.xml")

	// if err != nil {
	// 	panic(err)
	// }

	// data, err := io.ReadAll(f)

	// if err != nil {
	// 	panic(err)
	// }

	// users := new(Users)
	// err = xml.Unmarshal(data, users)

	// for _, user := range users.User {
	// 	t, _ := user.RegisteredToTime()

	// 	fmt.Println(t)
	// }
}
