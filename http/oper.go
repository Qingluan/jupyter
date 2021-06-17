package http

import (
	"log"
	"net/url"
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
	with.Article = NewArticle(with.Document)
	// with.Article = new(Article)
	// with.Each("title", func(i int, s *Selection) {
	// 	with.Article.Title = strings.TrimSpace(s.Text())
	// })
	// text := with.Document.Text()
	// var err error
	// for match, temp := range DateMatcher {
	// 	if res := match.FindAllString(text, -1); len(res) > 0 {
	// 		with.Article.Date, err = time.Parse(temp, res[0])
	// 		log.Println("parse time err:", err)
	// 		break
	// 	}
	// }
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
