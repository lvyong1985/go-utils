package go_utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Poem struct {
	Title     string
	TitleCode string
	Author    string
	Lines     []string
	LinesCode []string
}

func ServeWelcome(w http.ResponseWriter, welcomeHtml []byte, contextPath string) {
	welcome := string(welcomeHtml)

	poem, linesIndex := RandomPoem()

	welcome = strings.Replace(welcome, "<PoemTitle/>", poem.Title, 1)
	welcome = strings.Replace(welcome, "<PoemAuthor/>", poem.Author, 1)

	lines := ""
	for i, line := range poem.Lines {
		if i == linesIndex {
			lines += `<div style="color:red">` + line + `</div>`
		} else {
			lines += `<div>` + line + `</div>`
		}
	}

	welcome = strings.Replace(welcome, "<PoemLines/>", lines, 1)
	welcome = strings.Replace(welcome, `${/ContextPath}`, contextPath, -1)

	w.Write([]byte(welcome))
}

func RandomPoemBasicAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		basicAuthPrefix := "Basic "

		// 获取 request header
		auth := r.Header.Get("Authorization")
		// 如果是 http basic auth
		if strings.HasPrefix(auth, basicAuthPrefix) {
			// 解码认证信息
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)

				if len(pair) == 2 {
					user := string(pair[0])
					pass := string(pair[1])

					poem, linesIndex := RandomPoem()
					if user == poem.TitleCode && pass == poem.LinesCode[linesIndex] {
						fn(w, r) // 执行被装饰的函数
						return
					}
				}
			}
		}
		w.Header().Set("Content-Type", "'Content-type:text/html;charset=ISO-8859-1'")
		// 认证失败，提示 401 Unauthorized
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		// 401 状态码
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func RandomPoem() (Poem, int) {
	poems := ParsePoems("./poems.txt")
	now := time.Now()
	poemsIndex := now.Day() % len(poems)
	poem := poems[poemsIndex]
	linesIndex := int(now.Weekday()) % len(poem.LinesCode)
	return poem, linesIndex
}

func ParsePoems(poemFile string) []Poem {
	poems := make([]Poem, 0)
	poemsBytes, err := ioutil.ReadFile(poemFile)
	if err != nil {
		fmt.Println("read poems error", err.Error())
		return poems
	}

	fileLines := strings.Split(string(poemsBytes), "\n")

	for i := 0; i < len(fileLines); i++ {
		l := strings.TrimSpace(fileLines[i])

		if l == "" {
			continue
		}

		titleFields := strings.SplitN(l, "#", 2)
		i++
		author := strings.TrimSpace(fileLines[i])

		lines := make([]string, 0)
		linesCode := make([]string, 0)
		for i++; i < len(fileLines); i++ {
			if fileLines[i] == "" {
				break
			}

			lineFields := strings.SplitN(fileLines[i], "#", 2)
			lines = append(lines, lineFields[0])
			linesCode = append(linesCode, lineFields[1])
		}

		poems = append(poems, Poem{
			Title:     titleFields[0],
			TitleCode: titleFields[1],
			Author:    author,
			Lines:     lines,
			LinesCode: linesCode,
		})
	}

	return poems
}
