package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	NW = regexp.MustCompile(`\W`)

	DateMatcher = map[*regexp.Regexp]string{
		regexp.MustCompile(`[1-2]\d{3}年\d{1,2}月\d{1,2}日 \d{1,2}\:\d{1,2}\:\d{1,2}`): "2006年1月2日 15:04:05",
		regexp.MustCompile(`[1-2]\d{3}年[0-1]\d月[0-3]\d日 [0-2]\d\:[0-5]\d\:\d{2}`):   "2006年1月2日 15:04:05",
		regexp.MustCompile(`[1-2]\d{3}-\d{1,2}-\d{1,2} \d{1,2}\:\d{1,2}\:\d{1,2}`):  "2006-1-2 15:04:05",
		regexp.MustCompile(`[1-2]\d{3}/\d{1,2}/\d{1,2} \d{1,2}\:\d{1,2}\:\d{1,2}`):  "2006/1/2 15:04:05",
		regexp.MustCompile(`[1-2]\d{3}/\d{1,2}/\d{1,2} \d{1,2}\:\d{1,2}`):           "2006/1/2 15:04",
		regexp.MustCompile(`[1-2]\d{3}年\d{1,2}月\d{1,2}日 \d{1,2}\:\d{1,2}`):          "2006年1月2日 15:04",
		regexp.MustCompile(`[1-2]\d{3}-\d{1,2}-\d{1,2} \d{1,2}\:\d{1,2}`):           "2006-1-2 15:04",
		regexp.MustCompile(`[1-2]\d{3}年\d{1,2}月\d{1,2}日`):                           "2006年1月2日",
		regexp.MustCompile(`[1-2]\d{3}-\d{1,2}-\d{1,2}`):                            "2006-1-2",
		regexp.MustCompile(`[1-2]\d{3}/\d{1,2}/\d{1,2}`):                            "2006/1/2",
		regexp.MustCompile(`\w{1,15}, \d{1,2} \w{1,15} \d{4}`):                      "Mon, 02 Jan 2006",
		regexp.MustCompile(`\d{1,2} \w{1,15} \d{4}`):                                "02 Jan 2006",
		regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}\:\d{2}\:\d{2}Z`):                "2006-01-02T10:27:21Z",
		regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`):                                   "02.01.2006",
	}
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

func Skip(u string, us ...string) bool {
	for _, uu := range us {
		if uu == u {
			return true
		}
	}
	return false
}

func (with *WithOper) PreTestSkip(name string, urls ...string) (o []string) {
	name2 := filepath.Join(os.TempDir(), name)
	buf, err := ioutil.ReadFile(name2)
	if err != nil {
		Failed("no such cached!:", name2)
		ioutil.WriteFile(name2, []byte(""), os.ModePerm)
		// return async
		return
	}

	Success("Pre Test breakpoint :", name)
	e := make(map[string]int)
	for _, l := range strings.Split(string(buf), "\n") {
		ll := strings.TrimSpace(l)
		if strings.HasPrefix(ll, "http") {
			e[ll] = 0
		}
	}
	for _, u := range urls {
		if _, ok := e[u]; !ok {
			o = append(o, u)
		}
	}
	return

}

/** AsSiteMap 爬取site-map
提取xml sitemap的大部分标准

>@breakpointContinue 开启断点续传，会自动读取和存储 已爬页面到 /tmp/default-skip.txt 和 /tmp/skip-site.txt

>@showState 开启状态显示

>@filter 通过url 过滤每个channel true to entry false not entry
	example  func(u string){ return strings.Contains(u,"/zh/")}

*/
func (with *WithOper) AsSiteMap(do func(out *AsyncOut), breakpointContinue bool, showState bool, filter func(chanelUrl string) bool) *WithOper {
	name := filepath.Join(os.TempDir(), "skip-site.txt")

	done := map[string]int{}
	save := func(u string) {
		fp, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Println("cache tmp err:", err)
			return
		}
		defer fp.Close()
		fp.WriteString(u + "\n")
	}

	if breakpointContinue {
		buf, err := ioutil.ReadFile(name)
		if err != nil {
			log.Println("no cached site-map")
		} else {
			for _, l := range strings.Split(string(buf), "\n") {
				lll := strings.TrimSpace(l)
				if lll != "" && strings.HasPrefix(lll, "http") && strings.HasSuffix(lll, ".xml") {
					done[lll] = 1
				}
			}
		}
	}

	entrys := []string{}
	urls := []string{}
	// h, _ := with.Document.Html()
	// log.Println("do :", h)
	with.Each("loc", func(i int, s *Selection) {
		url := s.Text()
		if strings.HasSuffix(url, ".xml") {
			if _, ok := done[url]; !ok {
				if filter != nil {
					if filter(url) {
						entrys = append(entrys, url)
					}
				} else {
					entrys = append(entrys, url)
				}
				// save(url)
			} else {
				fmt.Println("\rSkip:", url)
			}

		} else {
			// log.Println(url)
			urls = append(urls, url)
			// asyncer.Async(url)
			// log.Println(url)
		}
	})
	// log.Println("1....")
	Success(Yello(" Channel :", len(entrys)))

	with.sess.Asyncs(7, true, showState, do, urls...)
	// log.Println("end")

	// urls = []string{}
	allcount := 0
	for {
		if len(entrys) == 0 {
			break
		}
		tmps := []string{}
		for _, url := range entrys {
			// Info("entry :", url)
			with.Entry(url).Each("loc", func(i int, s *Selection) {
				url := s.Text()

				if strings.HasSuffix(url, ".xml") {
					if _, ok := done[url]; !ok {
						if filter != nil {
							if filter(url) {
								tmps = append(tmps, url)
							}
						} else {
							tmps = append(tmps, url)
						}
					} else {
						fmt.Println("\rSkip:", url)
					}
				} else {
					// asyncer.Async(url)
					urls = append(urls, url)
				}
			})
			// Success("entry :", url)
			c := 0
			for {
				if len(urls) == 0 {
					if c > 10 {
						Failed("no news :", url)

						break
					}
					c += 1
					Failed("failed get news , wait ", c*10, " , try again ...")
					time.Sleep(time.Duration(c*10) * time.Second)
					with.Entry(url).Each("loc", func(i int, s *Selection) {
						url := s.Text()

						if strings.HasSuffix(url, ".xml") {
							if _, ok := done[url]; !ok {
								tmps = append(tmps, url)

							} else {
								fmt.Println("\rSkip:", url)
							}
						} else {
							// asyncer.Async(url)
							urls = append(urls, url)
						}
					})
				} else {
					break
				}
			}

			if len(urls) > 0 {
				// 预先测试是否存在跳过urls
				readyUrls := with.PreTestSkip(DEFAULT_BREAKPOINT_FILE, urls...)
				// 如果太多，再分片
				if len(readyUrls) > 300 {
					t := readyUrls[:300]
					readyUrls = readyUrls[300:]
					for {

						allcount += len(t)
						Success(Yello(" Current: ", len(t), " Left:", len(readyUrls), " Fin:", allcount), " Entry:", url)

						for i := 0; i < len(t)/20; i++ {
							L(" Wait : ", i, "/", len(t)/20, " s")
							time.Sleep(1 * time.Second)
						}
						with.sess.Asyncs(3, true, showState, do, t...)
						if len(readyUrls) < 300 {
							break
						}
						t = readyUrls[:300]
						readyUrls = readyUrls[300:]
					}
					Success(Yello(" Left:", len(readyUrls), " Fin:", allcount), " Entry:", url)

					for i := 0; i < len(readyUrls)/20; i++ {
						L(" Wait : ", i, "/", len(t)/20, " s")
						time.Sleep(1 * time.Second)
					}
					with.sess.Asyncs(3, true, showState, do, readyUrls...)

				} else {
					allcount += len(readyUrls)
					Success(Yello("Left: ", len(readyUrls), " Fin:", allcount), " Entry:", url)

					for i := 0; i < len(readyUrls)/20; i++ {
						L(" Left: ", len(readyUrls), "  Wait : ", i, "/", len(readyUrls)/20)
						time.Sleep(1 * time.Second)
					}
					with.sess.Asyncs(3, true, showState, do, readyUrls...)

				}
				save(url)

			}

			urls = []string{}
		}

		entrys = append([]string{}, tmps...)
		// }
	}
	// asyncer.EndAsync()
	// log.Println("end all")
	return with
}
