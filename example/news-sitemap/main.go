package main

import (
	"flag"
	"strings"

	"github.com/Qingluan/jupyter/http"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var myFlags arrayFlags

func main() {
	url := ""
	surl := ""
	proxy := ""
	save := ""
	// keys := []string{}
	flag.StringVar(&url, "u", "", "sitemap url")
	flag.StringVar(&surl, "S", "skip.txt", "sitemap url skip this")
	flag.StringVar(&proxy, "p", "socks5://127.0.0.1:1091", "sitemap url proxy")
	flag.StringVar(&save, "o", "out.json", "sitemap out.json")
	flag.Var(&myFlags, "key", "+ is must include , - is must not include")
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
		}, true, true, func(channelUrl string) bool {
			ff := true
			if len(myFlags) > 0 {
				ff = false
			}
			for _, k := range myFlags {
				if strings.HasPrefix(k, "+") {
					if strings.Contains(channelUrl, k[1:]) {
						ff = true
					}
				} else if strings.HasPrefix(k, "-") {
					if strings.Contains(channelUrl, k[1:]) {
						ff = false
					}
				}
			}
			return ff
		}).EndCache()
}
