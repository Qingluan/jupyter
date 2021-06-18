package http

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type WithOper struct {
	URL            *url.URL
	Document       *goquery.Document
	LastSelections []*goquery.Selection
	Links          Links
	Article        *Article
	Err            error
	sess           *Session
}

var (
	MsgCacheChan = make(chan string)
	CacheStatus  = false
)

type Article struct {
	Text   string    `json:"text"`
	Title  string    `json:"title"`
	Date   time.Time `json:"date"`
	Author string    `json:"author"`
}

var (
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
	}
)

func (with *WithOper) Each(css string, do ...func(i int, s *Selection)) *WithOper {
	with.Document.Find(css).Each(func(i int, a *goquery.Selection) {
		if do != nil {
			do[0](i, &Selection{*a})
		}
		with.LastSelections = append(with.LastSelections, a)
	})
	return with
}

func (with *WithOper) For(do func(i int, s *Selection)) *WithOper {
	for no, ele := range with.LastSelections {
		do(no, &Selection{*ele})
	}
	return with
}

func (with *WithOper) AsArticle() *WithOper {
	if ar := NewArticle(with.Document); ar != nil {
		// with.Articles = append(with.Articles, ar)
		with.Article = ar
	}
	return with
}

func (article *Article) WaitToFile() {
	// bufstr := ""
	// for article := range with.Articles {
	buf, err := json.Marshal(article)
	if err != nil {
		log.Println("ToFile Err:", err)
		// continue
		return
	}
	MsgCacheChan <- string(buf)
}

func (with *WithOper) StartCache(name string) *WithOper {
	if !CacheStatus {
		log.Println("start Cache:", name)
		CacheStatus = true
		go func(msg *chan string) {
			defer func() { CacheStatus = false }()

			bufstr := ""
			tick := time.NewTicker(time.Second * 3)
			for {
				select {
				case i := <-*msg:
					if i == "[END]" {
						break
					}

					bufstr += i + "\n"
				case <-tick.C:
					if bufstr != "" {
						fp, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
						if err != nil {
							log.Println("to file err :", err)
							break
						}
						fp.WriteString(bufstr)
						fp.Close()
						bufstr = ""
					}
				}

			}
		}(&MsgCacheChan)
	} else {
		log.Println("[stop old Cache]")
		MsgCacheChan <- "[END]"
		CacheStatus = false
		with.StartCache(name)
	}
	return with
}

func (with *WithOper) EndCache() *WithOper {
	MsgCacheChan <- "[END]"
	CacheStatus = false
	return with
}

func NewArticle(doc *goquery.Document) (article *Article) {
	article = new(Article)
	text1 := ""
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		article.Title = strings.TrimSpace(s.Text())
	})
	doc.Find("script").Remove()
	doc.Find("head").Remove()

	doc.Find("li").Remove()
	doc.Find("a").Remove()
	raw := doc.Text()
	// fmt.Println(raw)
	var err error
	for match, temp := range DateMatcher {
		if res := match.FindAllString(raw, -1); len(res) > 0 {
			article.Date, err = time.Parse(temp, res[0])
			if err != nil {
				log.Println("parse time err:", err)
			}
			break
		}
	}
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text1 += strings.TrimSpace(s.Text()) + "\n"
	})

	divClass := map[string]int{}
	max := ""
	maxI := 0
	doc.Find("div[class]").Each(func(i int, s *goquery.Selection) {
		n := s.AttrOr("class", "")

		if e, ok := divClass[n]; ok {
			e++
			if e > maxI {
				max = s.AttrOr("class", "")
				maxI = e
			}
			divClass[s.AttrOr("class", "")] = e
		} else {
			divClass[n] = 1

		}

	})
	text2 := ""
	css := "div[class=\"" + strings.TrimSpace(max) + "\"]"
	// log.Println("Found Css:", css)
	doc.Find(css).Each(func(i int, s *goquery.Selection) {
		// fmt.Println("Class:", s.Text())
		text2 += strings.TrimSpace(s.Text()) + "\n"
	})
	if len(text2) > len(text1) {
		article.Text = text2
	} else {
		article.Text = text1
	}
	return
}

func (with *WithOper) Entry(url string) *WithOper {
	return with.sess.With(url)
}
