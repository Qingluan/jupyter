package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Qingluan/jupyter/http"
)

func main() {
	url := ""
	surl := ""
	proxy := ""
	save := ""
	flag.StringVar(&url, "u", "", "sitemap url")
	flag.StringVar(&surl, "S", "skip.txt", "sitemap url skip this")
	flag.StringVar(&proxy, "p", "socks5://127.0.0.1:1091", "sitemap url proxy")
	flag.StringVar(&save, "o", "out.json", "sitemap out.json")

	flag.Parse()

	buf, err := ioutil.ReadFile(surl)
	skips := []string{}
	if err != nil {
		log.Println("err:", err)
	}
	for _, l := range strings.Split(string(buf), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		skips = append(skips, strings.TrimSpace(l))
	}
	session := http.NewSession()
	// pool := merkur.NewProxyPool(proxy)

	session.SetProxy(proxy)
	session.RandomeUA = true
	session.With(url).StartCache(save).
		AsSiteMap(func(with *http.AsyncOut) {
			with.Res.AsArticle()
			with.Res.Article.WaitToFile()
			// fmt.Print(".")
			// log.Println(" - ", with.Res.Article.Title)
		}, skips...).EndCache()
	// Each().
	// EndAsync()

	// with := session.With("https://cn.nytimes.com/world/20160728/t28terrorvictims/zh-hant/")
	// fmt.Println("pull")
	// article := with.AsArticle().Article
	// fmt.Println(article.Date.String())
	// fmt.Println(article.Title)
	// fmt.Println(article.Text)
}
