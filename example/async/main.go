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
	session.SetSocks5Proxy("127.0.0.1:1091")
	UA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36"
	session.SetHeader("User-agent", UA)
	// https://cn.nytstyle.com/international/20160728/mass-killings-contagion-copycat/zh-hant/
	// https://cn.nytimes.com/usa/20160726/cc26dnclive/zh-hant/
	// https://cn.nytstyle.com/international/20160728/dnc-biden-kaine-obama/zh-hant/
	// //

	session.With("https://www.google.com")
	fmt.Println("with ok")

	session.StartAsync(3).
		Async("https://cn.nytimes.com/china/20160729/china-beijing-traffic-pollution/").
		Async("https://cn.nytstyle.com/international/20160728/t28terrorvictims/").
		Async("https://cn.nytstyle.com/international/20160728/mass-killings-contagion-copycat/").
		Async("https://cn.nytimes.com/usa/20160726/cc26dnclive/").
		Async("https://cn.nytimes.com/china/20160728/china-yanhuang-chunqiu/").
		Async("https://cn.nytstyle.com/international/20160728/dnc-biden-kaine-obama/").
		Async("https://cn.nytstyle.com/international/20160729/north-korea-hacking-interpark/").
		Async("https://cn.nytimes.com/china/20160728/china-sex-workers-condoms-police/").
		Async("https://cn.nytimes.com/china/20160729/china-communist-party-propaganda/").
		Async("https://cn.nytstyle.com/health/20160727/alzheimers-checklist-mild-behavioral-impairment/").
		Async("https://cn.nytimes.com/china/20160708/nicholas-kristof-china-tiananmen-times-insider/").
		Async("https://cn.nytstyle.com/slideshow/20160719/t19rnc-ss/").
		Async("http://cn.nytimes.com/topic/20160712/archiev-topic/").
		Each(func(with *http.AsyncOut) {
			ar := with.Res.AsArticle().Article
			fmt.Println(ar.Title, fmt.Sprintf("[%s]", ar.Date.Local().String()))
		}).
		EndAsync()

	// with := session.With("https://cn.nytimes.com/world/20160728/t28terrorvictims/zh-hant/")
	// fmt.Println("pull")
	// article := with.AsArticle().Article
	// fmt.Println(article.Date.String())
	// fmt.Println(article.Title)
	// fmt.Println(article.Text)
}
