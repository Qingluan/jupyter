package main

import (
	"fmt"
	"os"

	"github.com/Qingluan/jupyter/http"
)

func main() {
	session := http.NewSession()
	// for _, link := range session.With(os.Args[1]).News().Links[0] {
	// 	log.Println(link)

	// }
	// _, err := session.Get("https://www.google.com")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("test ok")
	session.SetSocks5Proxy("127.0.0.1:1091")
	UA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"
	session.SetHeader("User-agent", UA)
	with := session.With(os.Args[1])
	fmt.Println("pull")
	article := with.AsArticle().Article
	fmt.Println(article.Date.String())
	fmt.Println(article.Title)
	fmt.Println(article.Text)
}
