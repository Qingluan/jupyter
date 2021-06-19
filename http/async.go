package http

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AsyncOut struct {
	Url string
	// Err error
	End bool
	Res *WithOper
}

type Async struct {
	num          int
	running      int
	Done         []string
	lock         sync.WaitGroup
	input        chan string
	out          chan AsyncOut
	runningState chan int
	session      *Session
	do           func(out *AsyncOut)
	save         func()
	// errRateInterval int
	// last
}

func (session *Session) StartAsync(i int) *Async {
	async := new(Async)
	async.input = make(chan string, i)
	async.out = make(chan AsyncOut, i)
	async.num = i
	async.session = session.Copy()
	async.session.SetTimeout(12)
	async.runningState = make(chan int, i*4)
	async.lock.Add(1)
	go async.run(&async.lock)

	return async
}

func (session *Session) Asyncs(work int, loadCache bool, showState bool, do func(each *AsyncOut), urls ...string) *Session {
	if urls != nil {
		asyn := session.StartAsync(work)
		if loadCache {
			asyn.LoadCache("default-skip.txt")
		}
		if showState {
			asyn.State()
		}
		asyn.Each(do)
		for _, url := range urls {
			// log.Println("u:", url)
			asyn.Async(url)
		}
		asyn.EndAsync()

	}
	return session
}

func (async *Async) Async(url string, proxy ...string) *Async {
	if Skip(url, async.Done...) {
		Info("skip:", url)
		return async
	}
	async.running++
	if proxy != nil && proxy[0] == "" {
		async.session.Proxy = ""
	}
	async.input <- url
	return async
}

func (async *Async) EndAsync() *Session {

	time.Sleep(1 * time.Second)
	async.input <- "[END]"
	// log.Println("Wait to End:", async.running)
	async.lock.Wait()
	return async.session
}

func (async *Async) Each(do func(out *AsyncOut)) *Async {
	async.do = do
	async.lock.Add(1)
	defer func() {
		if async.save != nil {
			async.save()
		}
	}()
	go func(cal func(res *AsyncOut), l *sync.WaitGroup, out chan AsyncOut, in chan string) {
		defer l.Done()
		errTrys := map[string]int{}
		// defer fmt.Println("End Each")
		time.Sleep(1 * time.Second)
		tick := time.NewTicker(10 * time.Second)
	E2:
		for {
			select {
			case outres := <-out:
				if outres.End {
					break E2
				}
				// log.Println("Out:", outres)
				if outres.Res.Err != nil {
					if e, ok := errTrys[outres.Url]; ok {
						if e > 3 {
							delete(errTrys, outres.Url)
							continue
						}
						errTrys[outres.Url] = e + 1
					} else {
						in <- outres.Url
						errTrys[outres.Url] = 1
					}
					continue
				}
				if outres.Url != "" {
					async.Done = append(async.Done, outres.Url)
				}
				cal(&outres)

			case <-tick.C:
				if async.save != nil {
					async.save()
				}
			}
			// *cal--

		}
	}(async.do, &async.lock, async.out, async.input)
	return async
}

func (async *Async) LoadCache(name string) *Async {
	name2 := filepath.Join(os.TempDir(), name)
	buf, err := ioutil.ReadFile(name2)
	if err != nil {
		Failed("no such cached!:", name2)
		ioutil.WriteFile(name2, []byte(""), os.ModePerm)
		return async
	}

	L("Load Cache :", name)
	for _, l := range strings.Split(string(buf), "\n") {
		ll := strings.TrimSpace(l)
		if strings.HasPrefix(ll, "http") {
			async.Done = append(async.Done, ll)

		}
	}

	async.save = func() {
		if len(async.Done) == 0 {
			return
		}
		us := async.Done

		name2 := filepath.Join(os.TempDir(), name)
		ioutil.WriteFile(name2, []byte(strings.Join(us, "\n")+"\n"), os.ModePerm)
		// fp, err := os.OpenFile(name2, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
		// if err != nil {
		// 	log.Println("cache tmp err:", err)
		// 	return
		// }
		// defer fp.Close()
		// fp.WriteString(strings.Join(us, "\n") + "\n")
		// async.Done = []string{}
	}
	return async
}
func (async *Async) State() *Async {
	L("show state")
	go func() {
		time.Sleep(3 * time.Second)

		for {
			c := len(async.runningState)
			ic := len(async.input)
			oc := len(async.out)
			if c == 0 && ic == 0 && oc == 0 {
				break
			}
			time.Sleep(4 * time.Second)
			L(Green("running : ", c), " ", Yello(" input : ", ic), Green(" output:", oc), Yello(" Done: ", len(async.Done)), " ")

		}
	}()

	return async
}
func (async *Async) run(as *sync.WaitGroup) {
	defer as.Done()
	// defer fmt.Println("End run")
	// c := 0
	outchan := async.out
	lock := sync.WaitGroup{}
	tick := time.NewTicker(10 * time.Second)
	Finish := 0
	// canEND := false
	// O:
	interval := time.Now().Add(5 * time.Second)
	errRate := 0
	// finishedUrls := []string{}

	balanceSleep := func() {
		if time.Now().After(interval) {
			errRate = 0
			interval = time.Now().Add(5 * time.Second)
			return
		}
		if errRate > 1 {
			defer Info("[wait unlock]")
			Info("[balance Hit may wait...", errRate, " s]")
			time.Sleep(time.Duration(errRate) * time.Second)
		}
	}
	for {
		inUrl := <-async.input
		// async.lock.Add(1)
		if inUrl == "[END]" {
			break
		}
		// log.Println("async:", inUrl)
		// c++
	E:
		for {

			c := len(async.runningState)

			select {
			case <-tick.C:
				// saveUrls(finishedUrls...)
			default:
				if c < async.num*4 {
					break E
				}
			}
			time.Sleep(100 * time.Microsecond)

		}

		balanceSleep()

		nsess := async.session.Copy()
		lock.Add(1)
		async.runningState <- 1
		go func(cond *sync.WaitGroup, out chan AsyncOut, running chan int, sess *Session, url string) {
			// defer
			// fmt.Println("async:", url, sess.Proxy)
			defer func() {
				<-running
				Finish++
				// L("\r[running : %d Finish: %d ] %s                \r", len(running), Finish, url)
				cond.Done()
			}()
			with := sess.With(url)

			if with.Err != nil {
				// d := time.Now().Sub(now)
				errRate += 1

			}

			res := AsyncOut{
				Url: url,
				// Err: with.Err
				Res: with,
			}
			out <- res

		}(&lock, outchan, async.runningState, nsess, inUrl)

	}
	Success("[asyncs loop end wait , finish :", Finish, "]")
	lock.Wait()
	async.out <- AsyncOut{End: true}
}
