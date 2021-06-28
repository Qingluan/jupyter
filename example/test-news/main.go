package main

import (
	"flag"

	"github.com/Qingluan/jupyter/http"
)

func main() {
	f := ""
	flag.StringVar(&f, "u", "", "target url")
	flag.Parse()

	sess := http.NewSession()
	sess.SetProxy("socks5://127.0.0.1:1091")
	for _, ls := range sess.With(f).SimpleNews().Links {
		for _, l := range ls {
			http.Success(l.GetTitle(), l.GetUrl())
		}
	}

}
