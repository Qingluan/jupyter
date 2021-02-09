package main

import (
	"flag"
	"fmt"

	"github.com/Qingluan/jupyter/http"
)

func main() {
	u := "https://news.163.com/"
	flag.StringVar(&u, "u", "https://news.163.com/", "news url")
	flag.Parse()
	sess := http.NewSession()
	res, _ := sess.Get(u)
	res.CssSelect("a[href]", func(i int, s *http.Selection) {
		// web.L("Level", fmt.Sprint(no), "-------------------------")
		// for _, l := range links {
		if i < 10 {
			fmt.Println(s.Text())
			// }
		}

	})
}
