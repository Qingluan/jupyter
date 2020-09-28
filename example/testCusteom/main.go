package main

import (
	"fmt"

	"github.com/Qingluan/jupyter/http"
	//	"github.com/Qingluan/merkur"
)

func main() {
	sess := http.NewSession()
	//http.SetProxyGenerater(merkur.NewProxyDialer)
	sess.GetsWith("http://ken737bn.win/?{id}", http.Gfunc{
		"id": func(last http.Value) http.Value {
			if last.Empty() {
				return http.NewValue(1)
			}
			if i, _ := last.AsInt(); i > 100 {
				return http.Value{}
			}
			return last.Increase()
		},
	}, func(loger http.Loger, res *http.SmartResponse, err error) {
		if out := res.CssExtract(http.Dict{
			"uid":         "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > h2 > span | text",
			"name":        "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > h2 | text ",
			"reply":       "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(1) > ul.cl.bbda.pbm.mbm > li > a:nth-child(4) | text",
			"group":       "#ct > div > div.bm.bw0 > div > div.bm_c.u_profile > div:nth-child(2) > ul:nth-child(2) > li:nth-child(2) > span > a | text",
			"regist time": "#pbbs > li:nth-child(2) | text",
		}); out != nil {
			fmt.Println(out)
		}
	}, "socks5://127.0.0.1:1091")
	// if resp, err := sess.Get("https://www.google.com", "socks5://127.0.0.1:1091"); err == nil {
	// 	fmt.Println(resp.Title())
	// }
}
