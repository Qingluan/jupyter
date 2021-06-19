package main

import (
	"flag"

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

	session := http.NewSession()

	session.SetProxy(proxy)
	session.RandomeUA = true
	session.With(url).
		// 开启缓存本地
		StartCache(save).
		// sitemap 提取模式
		AsSiteMap(func(with *http.AsyncOut) {
			// 提取新闻内容 ，发布时间 ，链接， 标题
			with.Res.AsArticle()
			// 扔进队列等待缓存到本地
			with.Res.Article.WaitToFile()
		}, true, true).EndCache()

}
