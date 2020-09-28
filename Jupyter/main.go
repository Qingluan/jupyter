package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Qingluan/jupyter/http"
	"github.com/Qingluan/merkur"
)

var (
	proxy        string
	target       string
	startId      int
	endId        int
	config       string
	output       string
	discuzzTempl bool
)

func main() {
	flag.StringVar(&config, "c", "", "config file")
	flag.BoolVar(&discuzzTempl, "H", false, "true  to show dz3.4 enumer user teplate  ")
	flag.Parse()
	extractD := make(http.Dict)
	if discuzzTempl {
		http.ShowDemo()
		os.Exit(0)
	}

	if config == "" {
		log.Println("need config file")
		os.Exit(1)
	}
	config := http.ReadConf(config)
	// fmt.Println(config)
	extractD = config.Template
	// }
	proxy = config.Proxy
	target = config.Domain
	startId = config.StartId
	endId = config.EndId
	names := config.Names
	output = config.Output
	// fmt.Println(names)
	// sort.Strings(names)
	if proxy != "" {
		// client := merkur.NewProxyHttpClient(proxy)
		// pool := merkur.NewProxyPool(proxy)
		http.SetProxyGenerater(merkur.NewProxyDialer)
		http.DefaultProxyPool = merkur.DefaultProxyPool

		// pool.Mode = merkur.Random
		http.LogLevl = 3
		sess := http.NewSession()
		sess.MultiGetRetryTime = 5
		sess.SetTimeout(7)

		suss_num := 0

		var x *csv.Writer
		if output != "" {
			fp, err := os.Create(output)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			x = csv.NewWriter(fp)
			// defer x.Close()
		}
		// sess.Get("https://")
		sess.RandomeUA = true
		sess.GetsWith(target, http.Gfunc{
			"id": func(last http.Value) http.Value {
				if last.Empty() {
					return http.NewValue(startId)
				}
				if i, _ := last.AsInt(); i > endId { //120303 {
					return http.Value{}
				}
				return last.Increase()
			},
		}, func(loger http.Loger, res *http.SmartResponse, err error) {
			if err == nil {
				// http.Success(suss_num, " | ", res.Text())

				if out := res.CssExtract(extractD); out != nil {
					if output != "" {
						values := []string{}
						skip := true
						for _, n := range names {
							// fmt.Println(n, out)
							vv := out[n]
							switch vv.(type) {
							case string:

								values = append(values, strings.TrimSpace(vv.(string)))
								if vv.(string) != "" {
									skip = false
								}
								// break
							case nil:
							}
						}
						if !skip {
							// fmt.Println(values)
							x.Write(values)
							x.Flush()
							http.Success(suss_num, res.RequestURL(), output)
						}
					} else {
						http.Success(suss_num, res.RequestURL())
					}
					suss_num++
				}
			} else {
				// http.Failed(res.RequestURL(), err)
			}
		}, proxy)
		// x.
	} else {
		sess := http.NewSession()
		sess.RandomeUA = true

		suss_num := 0
		var x *csv.Writer
		if output != "" {
			fp, err := os.Create(output)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			x = csv.NewWriter(fp)
			// defer x.Close()
		}

		sess.GetsWith(target, http.Gfunc{
			"id": func(last http.Value) http.Value {
				if last.Empty() {
					return http.NewValue(1)
				}
				if i, _ := last.AsInt(); i > 120303 {
					return http.Value{}
				}
				return last.Increase()
			},
		}, func(loger http.Loger, res *http.SmartResponse, err error) {
			if err == nil {
				// http.Success(suss_num, " | ", res.Text())

				if out := res.CssExtract(extractD); out != nil {
					if output != "" {
						values := []string{}
						skip := true
						for _, n := range names {
							vv := out[n].(string)
							values = append(values, vv)
							if vv != "" {
								skip = false
							}
						}
						if !skip {
							x.Write(values)
							x.Flush()
							http.Success(suss_num, res.RequestURL())
						}
					} else {
						http.Success(suss_num, res.RequestURL())
					}
					suss_num++
				}
			} else {
				// http.Failed(res.RequestURL(), err)
			}
		})
	}
}
