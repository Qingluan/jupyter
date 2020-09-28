package main

import (
	"fmt"

	"gitee.com/dark.H/Jupyter/http"
	"github.com/Qingluan/merkur"
)

func main() {
	sess := http.NewSession()
	http.SetProxyGenerater(merkur.NewProxyDialer)
	if resp, err := sess.Get("https://www.baidu.com"); err == nil {
		fmt.Println(resp.Title())
		for i, l := range resp.Soup().FindAll("a") {
			fmt.Println(i, l.Attrs()["href"], l.Text())
		}
		// fmt.Println()
	}

	if resp, err := sess.Get("https://www.google.com", "socks5://127.0.0.1:1091"); err == nil {
		fmt.Println(resp.Title())
	}
}
