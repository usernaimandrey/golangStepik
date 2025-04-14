package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

func easyjson9e1087fdDecodeHw3User(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = string(in.String())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9e1087fdEncodeHw3User(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeHw3User(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeHw3User(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeHw3User(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeHw3User(l, v)
}

const (
	android = "Android"
	msie    = "MSIE"
	// filePath = "./data/users.txt"
)

func FastSearch(out io.Writer) {
	f, err := os.Open("./data/users.txt")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	seenBrowsers := []string{}
	index := 0
	scanner := bufio.NewScanner(f)
	user := &User{}
	formatedEmail := ""

	out.Write([]byte("found users:\n"))

	for scanner.Scan() {
		byteText := scanner.Bytes()
		if !(strings.Contains(string(byteText), android) || strings.Contains(string(byteText), msie)) {
			index += 1
			continue
		}

		err := user.UnmarshalJSON(byteText)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {

			if strings.Contains(browser, android) {
				isAndroid = true
				if !slices.Contains(seenBrowsers, browser) {
					seenBrowsers = append(seenBrowsers, browser)
				}
			}

			if strings.Contains(browser, msie) {
				isMSIE = true
				if !slices.Contains(seenBrowsers, browser) {
					seenBrowsers = append(seenBrowsers, browser)
				}
			}
		}

		if !(isAndroid && isMSIE) {
			index += 1
			continue
		}

		buf := bytes.Buffer{}

		formatedEmail = strings.Replace(user.Email, "@", " [at] ", -1)

		// formatedUser := fmt.Sprintf("[%d] %s <%s>\n", index, user.Name, email)

		// out.Write([]byte(formatedUser))

		buf.WriteByte('[')
		buf.WriteString(strconv.Itoa(index))
		buf.WriteByte(']')
		buf.WriteByte(' ')
		buf.WriteString(user.Name)
		buf.WriteByte(' ')
		buf.WriteByte('<')
		buf.WriteString(formatedEmail)
		buf.WriteByte('>')
		buf.WriteByte('\n')
		out.Write(buf.Bytes())
		index += 1
	}
	out.Write([]byte("\n"))

	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func main() {
	fastOut := new(bytes.Buffer)
	FastSearch(fastOut)
	fastResult := fastOut.String()
	fmt.Println(fastResult)
}
