package http

import (
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	NW = regexp.MustCompile(`\W`)
)

type UrlSim struct {
	host    string
	title   string
	len     int
	w       int
	rank    int
	url     string
	no_w    []string
	structs []string
}

type Article struct {
	Text   string
	Title  string
	Date   time.Time
	Author string
}

func AsUrlSim(urlstr string, title ...string) (u *UrlSim) {

	u = new(UrlSim)
	u.url = urlstr
	u.len = len(urlstr)
	f, err := url.Parse(urlstr)
	if err != nil {
		log.Fatal("err url:", urlstr)
	}
	u.host = f.Host
	u.structs = NW.Split(urlstr, -1)
	u.no_w = NW.FindAllString(urlstr, -1)
	u.rank = len(u.structs)
	if title != nil {
		u.title = title[0]
	}
	return
}

func min(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func max(a, b interface{}) interface{} {
	switch a.(type) {
	case int:
		if a.(int) < b.(int) {
			return b
		} else {
			return a
		}
	case []interface{}:
		if len(a.([]interface{})) < len(b.([]interface{})) {
			return b
		} else {
			return a
		}
	default:
		panic("can not as compare")
	}
}

func AndSimHash(a, b string) float32 {
	if a == "" || b == "" {
		return 0
	}
	ml := min(len(a), len(b))
	max := max(len(a), len(b)).(int)
	core := 0
	for i := 0; i < ml; i++ {
		if a[i] != b[i] {
			break
		}
		core++
	}
	return float32(core) / float32(max)
}
func (u *UrlSim) SetTitle(title string) {
	u.title = title
}

func (u *UrlSim) GetTitle() string {
	return strings.TrimSpace(u.title)
}

func (u *UrlSim) GetUrl() string {
	return u.url
}

func (u *UrlSim) Sub(other interface{}) (score float32) {
	var u2 *UrlSim
	switch other.(type) {
	case *UrlSim:
		u2 = other.(*UrlSim)
	case string:
		u2 = AsUrlSim(other.(string))
	default:
		return -1
	}
	mm := min(u.len, u2.len)
	sam := 0
	for i := 0; i < mm; i++ {
		if u.url[i] != u2.url[i] {
			break
		}
		sam++
	}

	score = 1 - (float32(sam) / float32(max(u.len, u2.len).(int)))
	if score > 0.4 {
		mms := min(len(u.structs), len(u2.structs))
		var ssam float32
		for i := 0; i < mms; i++ {
			if u.structs[i] == u2.structs[i] {
				ssam += float32(1)
			} else {
				ssam += AndSimHash(u.structs[i], u2.structs[i])

			}
		}
		score = 1 - (float32(ssam) / float32(max(len(u.structs), len(u2.structs)).(int)))
	}
	return
}

type FilterOption struct {
	Rank     int
	Distance float32
	Proxy    interface{}
}
type Links [][]*UrlSim

func (link Links) AsString(rank int) (a [][]string) {

	for no, ls := range link {
		if no == rank {
			for _, l := range ls {
				aone := []string{l.GetTitle(), l.GetUrl()}
				a = append(a, aone)
			}
		}
	}
	return
}

func (with *WithOper) News(filters ...FilterOption) *WithOper {

	filter := FilterOption{
		Rank:     3,
		Distance: 0.3,
		Proxy:    nil,
	}
	if filters != nil {
		filter = filters[0]
	}
	links := Links{}
	urls := []*UrlSim{}
	basehost := with.URL.Hostname()
	urlsm := map[string]string{}
	with.Each("a[href]", func(i int, s *Selection) {
		if href, ok := s.Attr("href"); ok {
			t := s.Text()
			if len(t) < 7 {
				return
			}
			href = strings.TrimSpace(href)
			if strings.HasPrefix(href, "/") {
				href = filepath.Join(basehost, href)
				urlsm[href] = t
			} else if strings.HasPrefix(href, "http") {
				urlsm[href] = t
			} else if strings.HasPrefix(href, "javascript:;") {
			} else if strings.Contains(href, "javascript:void(0)") {
			} else {
				log.Println("pass", href, "pass")
			}
			// fmt.Println(s.Text())

		}

	})
	for i, v := range urlsm {
		urls = append(urls, AsUrlSim(i, v))
	}
	for no, ui := range urls {
		// ui := AsUrlSim(i)
		tmp := []*UrlSim{}
		for _, ui2 := range urls[no:] {
			if ui.url == ui2.url {
				continue
			}
			d := ui.Sub(ui2)
			if d < filter.Distance {
				tmp = append(tmp, ui2)
			}
		}
		if len(tmp) > 3 {
			links = append(links, tmp)
		}
	}
	sort.Slice(links, func(i, j int) bool {
		return len(links[i]) > len(links[j])
	})
	// sort.SliceSort(links, func(i, j int) bool {
	// 	return len(links[i]) > len(links[j])
	// })
	ff := min(filter.Rank, len(links))
	links = links[:ff]
	with.Links = links
	return with
}
