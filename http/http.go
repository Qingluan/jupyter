package http

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	httplib "net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb"

	// "github.com/roberson-io/mmh3"
	"github.com/reusee/mmh3"
	"golang.org/x/net/proxy"
)

var (
	UA = []string{
		"Mozilla/5.0 (Windows; U; Win98; en-US; rv:1.8.1) Gecko/20061010 Firefox/2.0",
		"Mozilla/5.0 (Windows; U; Windows NT 5.0; en-US) AppleWebKit/532.0 (KHTML, like Gecko) Chrome/3.0.195.6 Safari/532.0",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1 ; x64; en-US; rv:1.9.1b2pre) Gecko/20081026 Firefox/3.1b2pre",
		"Opera/10.60 (Windows NT 5.1; U; zh-cn) Presto/2.6.30 Version/10.60", "Opera/8.01 (J2ME/MIDP; Opera Mini/2.0.4062; en; U; ssr)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; ; rv:1.9.0.14) Gecko/2009082707 Firefox/3.0.14",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36",
		"Mozilla/5.0 (Windows; U; Windows NT 6.0; fr; rv:1.9.2.4) Gecko/20100523 Firefox/3.6.4 ( .NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.0; fr-FR) AppleWebKit/528.16 (KHTML, like Gecko) Version/4.0 Safari/528.16",
		"Mozilla/5.0 (Windows; U; Windows NT 6.0; fr-FR) AppleWebKit/533.18.1 (KHTML, like Gecko) Version/5.0.2 Safari/533.18.5",
	}
	// UA = random.choice(user_agent)
	DeafultHeaders = map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"User-Agent":                UA[0],
		"Upgrade-Insecure-Requests": "1",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Accept-Encoding":           "gzip, deflate, sdch",
		"Accept-Language":           "zh-CN,zh;q=0.8",
		"Referer":                   "http://www.baidu.com/link?url=www.so.com&url=www.soso.com&&url=www.sogou.com",
		"Cookie":                    "PHPSESSID=gljsd5c3ei5n813roo4878q203",
	}
)

type Session struct {
	Header            map[string]string
	Transprot         httplib.Transport
	Timeout           int
	RandomeUA         bool
	MultiGetRetryTime int
}

type SmartResponse struct {
	httplib.Response
	Code       int
	requestURL string
	cache      []byte
}

// type Document goquery.Document

func NewSession() (sess *Session) {
	sess = &Session{
		Header: DeafultHeaders,
		Transprot: httplib.Transport{
			ResponseHeaderTimeout: 22 * time.Second,
		},
		MultiGetRetryTime: 3,
		Timeout:           40,
	}
	return
}

func (session *Session) UrlJoin(f ...string) string {
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

func (session *Session) SetTimeout(t int) {
	session.Timeout = t
	session.Transprot.ResponseHeaderTimeout = time.Duration(t) * time.Second
}

func (session *Session) SetProxyDialer(dialer proxy.Dialer) {
	session.Transprot.Dial = dialer.Dial
}

func (session *Session) SetSocks5Proxy(proxyAddr string) (err error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return err
	}
	session.Transprot.Dial = dialer.Dial
	return nil
}

func (session *Session) SetHeader(key string, value string) {
	session.Header[key] = value
}

func Socks5Dialer(addr string) proxy.Dialer {
	if strings.HasPrefix(addr, "socks5") {
		addr = strings.SplitN(addr, "://", 2)[1]
	}
	dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		// os.Exit(1)
	}
	return dialer
}

func (session *Session) getClient(proxyObj ...interface{}) (client *httplib.Client) {
	transport := httplib.Transport{
		ResponseHeaderTimeout: session.Transprot.ResponseHeaderTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	if session.Transprot.Dial != nil {
		transport.Dial = session.Transprot.Dial
	}
	if proxyObj != nil && DefaultProxyDialer != nil {
		if proxyObj[0] == nil {
			log.Fatal("proxy is nil !!")
		}
		var dialer proxy.Dialer
		switch proxyObj[0].(type) {
		case proxy.Dialer:
			dialer = proxyObj[0].(proxy.Dialer)
		case string:
			if proxyObj[0].(string) != "" {
				if dialer = DefaultProxyDialer(proxyObj[0]); dialer == nil {
					log.Println("new proxy dialer create error:", proxyObj[0])
				}
			}
		case ProxyDiallerPool:
			ppol := proxyObj[0].(ProxyDiallerPool)
			dialer = ppol.GetDialer()
		case nil:
			panic("set proxy but null !")
		default:
			if dialer = DefaultProxyDialer(proxyObj[0]); dialer == nil {
				log.Fatal(proxyObj[0])
			}
		}
		if dialer == nil {
			return nil
		}

		transport.Dial = dialer.Dial

	} else {
		// log.Println("empty proxy ")
	}
	client = &httplib.Client{
		Transport: &transport,
		Timeout:   time.Duration(session.Timeout) * time.Second,
	}
	return
}

/**
* Get
	set proxy:
		socks5://xxx.x.x.x.x:port
		ss://xxasfsfs
		ssr://xasfsaf
		General.Config{...}
*/
func (session *Session) Get(url string, proxy ...interface{}) (resp *SmartResponse, err error) {
	req, err := httplib.NewRequest("GET", url, nil)
	if err != nil {
		// Failed("make req err:", err)
		return nil, err
	}
	for k, v := range session.Header {
		req.Header.Set(k, v)
	}
	client := session.getClient(proxy...)
	if client == nil {
		return nil, fmt.Errorf("Proxy Set Error: %v", proxy)
	}
	if session.RandomeUA {
		ix := rand.Int() % len(UA)
		ua := UA[ix]
		req.Header.Set("User-agent", ua)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp = &SmartResponse{
		*res,
		res.StatusCode,
		url,
		nil,
	}
	// client.CloseIdleConnections()
	return
}

func (session *Session) Post(httpurl string, data map[string]string, proxy ...interface{}) (resp *SmartResponse, err error) {
	u := url.Values{}
	for k, v := range data {
		u.Add(k, v)
	}

	req, err := httplib.NewRequest("POST", httpurl, strings.NewReader(u.Encode()))
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	client := session.getClient(proxy...)
	if client == nil {
		return nil, fmt.Errorf("Proxy Set Error: %v", proxy)
	}
	for k, v := range session.Header {
		req.Header.Set(k, v)
	}
	if session.RandomeUA {
		ix := rand.Int() % len(UA)
		ua := UA[ix]
		req.Header.Set("User-agent", ua)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp = &SmartResponse{
		*res,
		res.StatusCode,
		httpurl,
		nil,
	}
	return
}

func (session *Session) Json(url string, data map[string]string, proxy ...interface{}) (resp *SmartResponse, err error) {
	buf, _ := json.MarshalIndent(data, "", "\t")
	req, err := httplib.NewRequest("POST", url, bytes.NewBuffer(buf))
	req.Header.Set("Content-type", "application/json")

	client := session.getClient(proxy...)
	if client == nil {
		return nil, fmt.Errorf("Proxy Set Error: %v", proxy)
	}
	for k, v := range session.Header {
		req.Header.Set(k, v)
	}
	if session.RandomeUA {
		ix := rand.Int() % len(UA)
		ua := UA[ix]
		req.Header.Set("User-agent", ua)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp = &SmartResponse{
		*res,
		res.StatusCode,
		url,
		nil,
	}
	return
}

func (session *Session) Upload(url string, filePath string, fileKey string, data map[string]string, showBar bool, proxy ...interface{}) (resp *SmartResponse, err error) {

	fp, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	var fi os.FileInfo
	if fi, err = fp.Stat(); err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)
	var bar *pb.ProgressBar
	if showBar {
		bar = pb.New64(fi.Size()).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
		bar.Start()
	}
	go func() {
		var part io.Writer
		defer w.Close()
		defer fp.Close()

		for k, v := range data {
			w1, _ := mpw.CreateFormField(k)
			w1.Write([]byte(v))
		}

		if part, err = mpw.CreateFormFile(fileKey, fi.Name()); err != nil {
			log.Fatal(err)
		}
		if showBar {
			part = io.MultiWriter(part, bar)
		}
		if _, err = io.Copy(part, fp); err != nil {
			log.Fatal(err)
		}
		if err = mpw.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	req, err := httplib.NewRequest("POST", url, r)
	req.Header.Set("Content-type", mpw.FormDataContentType())

	client := session.getClient(proxy...)
	if client == nil {
		return nil, fmt.Errorf("Proxy Set Error: %v", proxy)
	}
	for k, v := range session.Header {
		req.Header.Set(k, v)
	}
	if session.RandomeUA {
		ix := rand.Int() % len(UA)
		ua := UA[ix]
		req.Header.Set("User-agent", ua)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp = &SmartResponse{
		*res,
		res.StatusCode,
		url,
		nil,
	}
	return
}

// ---------------------------- parse body --------------------------------------

func (smartres *SmartResponse) Html() []byte {
	if smartres.cache == nil {
		switch smartres.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err := gzip.NewReader(smartres.Body)
			if err != nil {
				log.Println("parse body gzip data error:", err)
				return nil
			}
			defer reader.Close()
			smartres.cache, _ = ioutil.ReadAll(reader)
		default:
			t, _ := ioutil.ReadAll(smartres.Body)
			smartres.cache = t
		}
	}

	return smartres.cache
}

func (smartres *SmartResponse) RequestURL() *url.URL {
	if smartres == nil {
		return nil
	}
	c, _ := url.Parse(smartres.requestURL)
	return c
}

func (smartres *SmartResponse) String() string {
	return string(smartres.Html())
}

func (smartres *SmartResponse) Json(obj ...interface{}) (jdata map[string]interface{}) {
	jdata = make(map[string]interface{})
	if obj != nil {
		json.Unmarshal(smartres.Html(), &obj[0])
	}
	err := json.Unmarshal(smartres.Html(), &jdata)
	if err != nil {
		jdata["output"] = err.Error()
	}
	return
}

func (smartres *SmartResponse) Search(key string, toLower bool) bool {
	if toLower {
		return strings.Contains(strings.ToLower(string(smartres.Html())), key)
	}
	return bytes.Contains(smartres.Html(), []byte(key))
	// return false
}

func (smartres *SmartResponse) HeaderJson() string {
	d := make(map[string]string)
	for k, v := range smartres.Header {
		d[k] = v[0]
	}
	buf, _ := json.MarshalIndent(d, "", "\t")
	return string(buf)
}

func (res *SmartResponse) Base64() string {
	return base64.StdEncoding.EncodeToString(res.Html())
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func (res *SmartResponse) Base64Mime() []byte {
	buffer := bytes.NewBuffer([]byte{})
	for _, sepbuf := range split(res.Html(), 57) {
		buf := make([]byte, 77)
		base64.StdEncoding.Encode(buf, sepbuf)
		ix := bytes.IndexByte(buf, 0x00) + 1
		buf = buf[:ix]
		buf[ix-1] = byte('\n')
		buffer.Write(buf)
	}
	return buffer.Bytes()
}

func (res *SmartResponse) HashMMH3Base64() int32 {
	key := []byte(res.Base64Mime())
	if len(key) > 28 {
		locker.Lock()
		// defer locker.Unlock()
		// fmt.Println(len(key))
		// var seed uint32 = 0
		return int32(mmh3.Hash32(key))
		// return int32(binary.LittleEndian.Uint32(hash))
	}
	return 0
}

func (res *SmartResponse) HashMMH3() int32 {
	key := res.Html()
	if len(key) > 28 {
		// var seed uint32 = 0
		return int32(mmh3.Hash32(key))
		// return int32(binary.LittleEndian.Uint32(hash))
	}
	return 0
}

func (smartres *SmartResponse) HeaderString() (d string) {
	for k, v := range smartres.Header {
		d += fmt.Sprintf("%s : %s\n", k, v[0])
	}
	// buf, _ := json.Marshal(d)
	return strings.TrimSpace(d)
}

// Get Soup
func (smartres *SmartResponse) Soup() (m *goquery.Document) {
	pre := bytes.NewBuffer(smartres.Html())
	ebuffer, err := DecodeHTMLBody(pre, "")
	if err != nil {
		log.Fatal("can not decode html")
	}
	d, e := goquery.NewDocumentFromReader(ebuffer)
	if e != nil {
		return nil
	} else {
		return d
	}
}

// Get title
func (smartres *SmartResponse) Title() string {
	if soup := smartres.Soup(); soup != nil {
		return soup.Find("title").Text()
	}
	return ""
}

// Get regex group
func (smartres *SmartResponse) ReExtractString(re string) []string {
	compile := regexp.MustCompile(re)
	// Failed("re compile err: ", err)
	return compile.FindStringSubmatch(string(smartres.Html()))
	// return []string{}
}

// get content md5
func (smartres *SmartResponse) Md5() string {
	return GetMD5(smartres.Html())
}

func GetMD5(c []byte) string {
	res := md5.Sum(c)
	m := hex.EncodeToString(res[:])
	// fmt.Println(">> ", m, " <<<")
	return m
}
