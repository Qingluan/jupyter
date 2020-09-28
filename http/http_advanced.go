package http

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

var (
	MAX_POOL   = 100
	RandomMode = 1
	FlowMode   = 0
	HostParse  = regexp.MustCompile(`Host: .+`)
	// RawReqStringSmartRe = regexp.MustCompile(`\n`)
)

// func (d *Document) Find(cssselct string) *Selection {
// 	rd := (*goquery.Document)(d)
// 	// rd.Each(cssselct, func(i int, d *goquery.Selection) {
// 	// a(i, (*Selection)(d))
// 	// })
// 	return &Selection{*rd.Find(cssselct)}
// 	// return *ff}
// }

// func (s *Selection) Each(deal func(i int, d *Selection)) *Selection {
// 	// rs := (*goquery.Selection)(s)
// 	return rs.Each(func(i int, a *goquery.Selection) {
// 		deal(i, (*Selection)(a))
// 	})
// }

func (sess *Session) MultiGet(urls []string, handleRes func(loger Loger, res *SmartResponse, err error), showBar bool, proxy ...interface{}) {
	c := len(urls)
	if c > 100 {
		c = 100
	}
	pool := NewAwaitPool(c)
	pool.RetryTime = sess.MultiGetRetryTime
	if proxy != nil {
		switch proxy[0].(type) {
		case string:
			if strings.HasPrefix(proxy[0].(string), "http") {
				InitProxyPool(proxy[0].(string))
			}
		}
	}
	pool.Handle = func(url string) interface{} {
		res, err := sess.Get(url, proxy...)
		if err != nil {
			return err
		}
		return res
	}

	// pool.ErrDo = func(err error,errorTimes int, res Result, loger Loger) {
	// 	if errTryTimes < retry{
	// 		pool.TryAgain()
	// 	}
	// }

	pool.After = func(res Result, loger Loger) {
		switch res.Res.(type) {
		case error:
			handleRes(loger, nil, fmt.Errorf("%s : %s  (proxy : %v)", res.Url, res.Res.(error).Error(), proxy))
		case *SmartResponse:
			defer res.Res.(*SmartResponse).Body.Close()
			handleRes(loger, res.Res.(*SmartResponse), nil)
		}
	}
	pool.Loop(urls, showBar)
}

func (session *Session) TestErrorPage(url string, proxy ...interface{}) (string, string, string) {
	MustErrorUrl := UrlJoin(url, "asdfiiangidsnignaidsgasdgisdanisndcinsadigsdasg_fu2k_you_b0y_t3st404")
	if res, err := session.Get(url, proxy...); err == nil {

		mainPage := res.PageTextHash()
		arbitraryId := mainPage
		if links := res.Links(); len(links) > 0 {
			if ls := ArrayFilter(links).FilterFunc(func(i int, s string) bool {
				if strings.Contains(s, "=") {
					return true
				}
				return false
			}); len(ls) > 0 {
				if res, err := session.Get(ls[0]+"arbitraryId", proxy...); err == nil {
					arbitraryId = res.PageTextHash()
				}
			}

		}

		if res, err := session.Get(MustErrorUrl, proxy...); err == nil {
			return mainPage, res.PageTextHash(), arbitraryId
		}
	}
	return "", "", ""
}

func (session *Session) Send(raw string, proxy ...interface{}) (resp *SmartResponse, err error) {
	var reader *bufio.Reader
	var host string
	if strings.Contains(raw, "\n\n") {
		heeader := strings.SplitN(raw, "\n\n", 2)
		heeader[0] = strings.ReplaceAll(heeader[0], "\n", "\r\n")
		reader = bufio.NewReader(strings.NewReader((heeader[0] + "\r\n\r\n" + heeader[1])))
	} else {
		reader = bufio.NewReader(strings.NewReader(strings.ReplaceAll(strings.TrimSpace(raw)+"\n\n", "\n", "\r\n")))
	}
	if fs := strings.SplitN(HostParse.FindString(raw), ":", 2); len(fs) > 1 {
		host = strings.TrimSpace(fs[1])
		log.Println("host:", host)
	}
	req, err := http.ReadRequest(reader)
	if req.Header.Get("Host") == "" {
		req.Host = host
		req.URL.Scheme = "http"
		req.URL.Host = host
		// req.RequestURI = req.URL.String()
		req.Header["Host"] = []string{host}
	}
	if err != nil {
		log.Println("parse err:", err)
		return
	} else {
		// log.Println("url:", req.URL)
	}
	reqNew, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	for k, v := range req.Header {
		for _, value := range v {
			reqNew.Header.Add(k, value)
		}
	}

	if client := session.getClient(proxy...); client != nil {
		if session.RandomeUA {
			ix := rand.Int() % len(UA)
			ua := UA[ix]
			reqNew.Header.Set("User-agent", ua)
			// req.Header.Set("User-agent", ua)
		}

		res, err := client.Do(reqNew)
		// res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		resp = &SmartResponse{
			*res,
			res.StatusCode,
			reqNew.URL.String(),
			// req.URL.String(),
			nil,
		}

	} else {
		err = fmt.Errorf("proxy create error: %v", proxy)
	}
	return
}

func (sess *Session) CheckAlive(urls []string, showBar bool, after func(res *SmartResponse) bool, proxy ...interface{}) (alived []string) {
	sess.MultiGet(urls, func(loger Loger, res *SmartResponse, err error) {
		if err == nil {
			if res.StatusCode/200 == 1 {
				if after != nil {
					if after(res) {
						alived = append(alived, res.RequestURL().String())
					}
				} else {

					alived = append(alived, res.RequestURL().String())

				}
			}
		}
	}, showBar, proxy...)
	return
}

func InitProxyPool(url string) {
	DefaultProxyPool.Add(url)
}

// 0 is flow / 1 is random
func DefaultProxyMode(mode int) {
	DefaultProxyPool.SetMode(mode)
}
