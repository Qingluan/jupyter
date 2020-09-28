package http

import (
	"bufio"
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	B2S                 = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	DelStyle            = regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
	DelScript           = regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	DelHtmlTag          = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	DelSpaceContinuesly = regexp.MustCompile("\\s{2,}")
	FileTp              = regexp.MustCompile(`\.[a-zA-Z0-9]+`)
)

func (res *SmartResponse) Links(includeouter ...bool) (s []string) {
	u := res.RequestURL()
	host := u.Host
	base := u.Scheme + "://" + host
	e := make(map[string]bool)
	res.Soup().Find("a").Each(func(i int, d *goquery.Selection) {

		link, ok := d.Attr("href")
		if !ok {
			return
		}
		if link == "/" || link == "" {
			return
		}
		if strings.HasPrefix(link, "/") {
			e[UrlJoin(base, link)] = true
		} else if strings.HasPrefix(link, "http") {
			if includeouter != nil && includeouter[0] {
				e[link] = true
			}
		}

	})

	return DictBool(e).Keys()
}

type Selection struct {
	goquery.Selection
}

func (res *SmartResponse) CssSelect(css string, each func(i int, s *Selection)) {
	res.Soup().Find(css).Each(func(i int, a *goquery.Selection) {
		each(i, &Selection{*a})
	})
}

func (res *SmartResponse) FileLinks(includeouter ...bool) (s []string) {
	e := map[string]bool{
		".txt":  true,
		".zip":  true,
		".m3u":  true,
		".jpg":  true,
		".png":  true,
		".ico":  true,
		".rar":  true,
		".conf": true,
		".md":   true,
		".git":  true,
	}
	return ArrayFilter(res.Links()).FilterFunc(func(i int, v string) bool {
		u, _ := url.Parse(v)
		if t := FileTp.FindString(u.Path); t != "" {
			if e[t] {
				return true
			}
		}
		return false
	})
}

func UrlJoin(f ...string) string {
	if len(f) == 0 {
		return ""
	}
	base := strings.TrimSpace(f[0])
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	for _, i := range f[1:] {
		if strings.HasPrefix(strings.TrimSpace(i), "/") {
			base += i
		} else {
			base += "/" + i
		}
	}
	return base
}

func (res *SmartResponse) PageTextHash() string {
	if buf := res.Text(); len(buf) > 0 {
		// Success("id:", buf)
		return GetMD5([]byte(buf))
	} else {
		return ""
	}
}

func (resp *SmartResponse) Text() string {
	src := string(resp.Html())

	//将HTML标签全转换成小写
	// re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = B2S.ReplaceAllStringFunc(src, strings.ToLower)

	//去除STYLE
	src = DelStyle.ReplaceAllString(src, "")

	//去除SCRIPT
	src = DelScript.ReplaceAllString(src, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	src = DelHtmlTag.ReplaceAllString(src, "\n")

	//去除连续的换行符
	src = DelSpaceContinuesly.ReplaceAllString(src, "\n")
	return strings.TrimSpace(src)
}

func ParseRawData(buf []byte, url string) (r *SmartResponse, err error) {
	buffer := bufio.NewReader(bytes.NewBuffer(buf))
	res, err := http.ReadResponse(buffer, nil)
	if err != nil {
		return nil, err
	}
	return &SmartResponse{
		*res,
		res.StatusCode,
		url,
		nil,
	}, nil
}

func (res *SmartResponse) FastCheckLineByLine(found func(line string) bool) (string, bool) {
	if res != nil {
		defer res.Body.Close()
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			l := scanner.Text()
			if found(l) {
				return l, true
			}
		}
	}
	return "", false
}
