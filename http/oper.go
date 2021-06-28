package http

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
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
	lock           sync.WaitGroup
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
	Link   string    `json:"link"`
}

var ()

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
	if with.Document != nil {
		if ar := NewArticle(with.Document); ar != nil {
			// with.Articles = append(with.Articles, ar)
			ar.Link = with.URL.String()
			with.Article = ar

		}
	}
	return with
}

func (article *Article) WaitToFile() {
	if article == nil {
		return
	}
	if article.Text == "" {
		return
	}
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

		CacheStatus = true
		with.lock.Add(1)
		go func(msg *chan string, l *sync.WaitGroup) {
			defer l.Done()
			defer func() { CacheStatus = false }()
			defer Failed("End Cache")
			Success("start Cache:", name)
			bufstr := ""
			tick := time.NewTicker(time.Second * 3)
		E:
			for {
				select {
				case i := <-*msg:
					if i == "[END]" {
						break E
					}
					// log.Println("[FILE] :", "<-", i)
					bufstr += i + "\n"
				case <-tick.C:
					if bufstr != "" {
						// log.Println("[SAVE] :", "->", name)

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
			if bufstr != "" {
				// log.Println("[SAVE] :", "->", name)

				fp, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
				if err != nil {
					log.Println("to file err :", err)
					// break
				}
				fp.WriteString(bufstr)
				fp.Close()
				bufstr = ""
			}
		}(&MsgCacheChan, &with.lock)
	} else {
		Success("[stop old Cache]")
		MsgCacheChan <- "[END]"
		CacheStatus = false
		with.StartCache(name)
	}
	return with
}

func (with *WithOper) EndCache() *WithOper {
	MsgCacheChan <- "[END]"
	CacheStatus = false
	with.lock.Wait()
	return with
}

func FindMostMayDate(title, raw string) (t time.Time) {
	// var err error
	ts := []time.Time{}
	for match, temp := range DateMatcher {
		if res := match.FindAllString(raw, -1); len(res) > 0 {
			t, err := time.Parse(temp, res[0])
			if err != nil {
				log.Println("parse time err:", err)
				continue
			} else {
				// fmt.Println("Date:", temp, res[0])
				ts = append(ts, t)
			}
			// break
		}
	}
	// now := time.Now()
	if len(ts) > 0 {
		min := ts[0]
		for _, tt := range ts {
			if tt.After(min) {
				min = tt
			}
		}
		return min
	} else {
		// log.Println("not found : use init")
		L(Red("Not found date for:", title))
		// return nil
		return time.Time{}
	}

}

func NewArticle(doc *goquery.Document) (article *Article) {
	article = new(Article)
	text1 := ""
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		article.Title = strings.TrimSpace(s.Text())
	})
	doc.Find("script").Remove()
	doc.Find("head").Remove()
	// doc.Find("li").Remove()
	doc.Find("a").Remove()
	raw := doc.Text()
	// fmt.Println(raw)
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
	if strings.TrimSpace(article.Text) == "" || len(strings.TrimSpace(article.Text)) < 30 {
		article.Text = ""
		return
	}
	article.Date = FindMostMayDate(article.Title+"|"+article.Link, raw)

	return
}

func (with *WithOper) Entry(url string) *WithOper {
	return with.sess.With(url)
}
