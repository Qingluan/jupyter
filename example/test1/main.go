package main

import (
	"fmt"

	"github.com/Qingluan/jupyter/http"
	//	"github.com/Qingluan/merkur"
)

func main() {
	sess := http.NewSession()
	//http.SetProxyGenerater(merkur.NewProxyDialer)
	if resp, err := sess.Get("https://www.baidu.com"); err == nil {
		fmt.Println(resp.Title())
		resp.CssSelect("a", func(i int, l *http.Selection) {
			http.L(i, " ", l.Text(), " ", l.AttrOr("href", "X"))
		})
		// fmt.Println()
	}

	// if resp, err := sess.Get("https://www.google.com", "socks5://127.0.0.1:1091"); err == nil {
	// 	fmt.Println(resp.Title())
	// }
}
