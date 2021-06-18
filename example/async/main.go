package main

import (
	"fmt"

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
	session.SetProxy("socks5://127.0.0.1:1091")
	UA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"
	session.SetHeader("User-agent", UA)
	// https://cn.nytstyle.com/international/20160728/mass-killings-contagion-copycat/zh-hant/
	// https://cn.nytimes.com/usa/20160726/cc26dnclive/zh-hant/
	// https://cn.nytstyle.com/international/20160728/dnc-biden-kaine-obama/zh-hant/
	// //

	session.With("https://www.bbc.com/sitemaps/https-sitemap-com-archive-2.xml").StartCache("bbc.json").
		AsSiteMap(func(with *http.AsyncOut) {
			with.Res.AsArticle()
			with.Res.Article.WaitToFile()
			fmt.Print(".")
		}).EndCache()
	// Each().
	// EndAsync()

	// with := session.With("https://cn.nytimes.com/world/20160728/t28terrorvictims/zh-hant/")
	// fmt.Println("pull")
	// article := with.AsArticle().Article
	// fmt.Println(article.Date.String())
	// fmt.Println(article.Title)
	// fmt.Println(article.Text)
}
