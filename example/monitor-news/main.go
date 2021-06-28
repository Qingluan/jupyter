package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/jupyter/http"
)

var (
	GNews = make(map[string]string)
	lock  = sync.Mutex{}
)

func PushDing(msg, dingu string) {

	// url = "https://oapi.dingtalk.com/robot/send?access_token=cc7b356c8a88d5e0c9b7c02d6ac68aaf9c903a6f83f465ef80e77de59f9a8c3d"
	url := dingu
	values := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("Hit It:%s", msg)},
		"at": map[string]interface{}{
			"atMobiles": []string{"@所有人"},
		},
	}
	// headers = {'Content-Type': 'application/json; charset=UTF-8'}
	sess := http.NewSession()
	sess.Json(url, values)

}

func Save(tmpName string) {

	buf, err := json.Marshal(GNews)
	if err != nil {
		http.Failed("save to ", tmpName, " Failed with ", err)
		return
	}
	ioutil.WriteFile(tmpName, buf, os.ModePerm)
}

func Load(tmpName string) {
	lock.Lock()
	defer lock.Unlock()
	buf, err := ioutil.ReadFile(tmpName)
	if err != nil {
		http.Failed("load ", tmpName, " Failed with ", err)
		return
	}
	json.Unmarshal(buf, &GNews)
}

func Done(uf, robot, proxy string) {
	// lock.Lock()
	buf, err := ioutil.ReadFile(uf)
	if err != nil {
		http.Failed("read file:", err)

	}
	ls := []string{}
	for _, l := range strings.Split(string(buf), "\n") {
		ll := strings.TrimSpace(l)
		if !strings.HasPrefix(ll, "http") {
			ll = "http://" + ll
		}
		ls = append(ls, ll)
	}
	sess := http.NewSession()
	sess.SetProxy(proxy)
	sess.RandomeUA = true
	for {

		if len(ls) < 100 {
			break
		}
		cs := ls[:100]
		ls = ls[100:]
		sess.Asyncs(5, false, true, func(a *http.AsyncOut) {
			all := ""
			if len(a.Res.News().Links) > 0 {
				for _, link := range a.Res.News().Links[0] {
					if _, ok := GNews[link.GetTitle()]; !ok {
						title := link.GetTitle()
						link := link.GetUrl()
						GNews[title] = link
						all += fmt.Sprintf("%s ( %s ) \n", title, link)
					}
				}
			}
			PushDing(all, robot)
		}, cs...)

	}
	sess.Asyncs(5, false, true, func(a *http.AsyncOut) {
		all := ""
		if len(a.Res.News().Links) > 0 {
			for _, link := range a.Res.SimpleNews().Links[0] {
				if _, ok := GNews[link.GetTitle()]; !ok {
					title := link.GetTitle()
					link := link.GetUrl()
					GNews[title] = link
					all += fmt.Sprintf("%s ( %s ) \n", title, link)
				}
			}
		}
		if all != "" {
			PushDing(all+"\n 网站："+a.Url, robot)
		}
	}, ls...)
}

func main() {
	urlFiles := ""
	robotUrl := ""
	proxy := ""
	saveFile := ""
	flag.StringVar(&urlFiles, "u", "urls.txt", "set url to read")
	flag.StringVar(&robotUrl, "r", "https://oapi.dingtalk.com/robot/send?access_token=e50fbac84694b9014b254b020717d6fd8bff0430f70129b30025748d0ffebfed", "robot urls")
	flag.StringVar(&proxy, "p", "socks5://127.0.0.1:1091", "set proxy ")
	flag.StringVar(&saveFile, "s", "save.cache", "tmp news url cached")
	flag.Parse()
	Load(saveFile)
	min60 := time.NewTicker(60 * time.Minute)

	Done(urlFiles, robotUrl, proxy)
	Save(saveFile)
	for {
		select {
		case <-min60.C:

			Done(urlFiles, robotUrl, proxy)
			Save(saveFile)
		default:
			time.Sleep(1 * time.Second)
			http.L(http.Green(time.Now()), http.Yello(" Now : ", len(GNews)))
		}
	}
}
