package http

import (
	"fmt"
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
	lock         sync.WaitGroup
	input        chan string
	out          chan AsyncOut
	runningState chan int
	session      *Session
	do           func(out *AsyncOut)
}

func (session *Session) StartAsync(i int) *Async {
	async := new(Async)
	async.input = make(chan string, i)
	async.out = make(chan AsyncOut, i)
	async.num = i
	async.session = session.Copy()
	async.runningState = make(chan int, i*4)
	async.lock.Add(1)
	go async.run(&async.lock)

	return async
}

func (session *Session) Asyncs(work int, do func(each *AsyncOut), urls ...string) *Session {
	if urls != nil {
		asyn := session.StartAsync(work).Each(do)
		for _, url := range urls {
			// log.Println("u:", url)
			asyn.Async(url)
		}
		asyn.EndAsync()

	}
	return session
}

func (async *Async) Async(url string, proxy ...string) *Async {
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
	go func(cal func(res *AsyncOut), l *sync.WaitGroup, out chan AsyncOut, in chan string) {
		defer l.Done()
		errTrys := map[string]int{}
		// defer fmt.Println("End Each")
		time.Sleep(1 * time.Second)
		for {
			outres := <-out
			if outres.End {
				break
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
			}

			cal(&outres)
			// *cal--

		}
	}(async.do, &async.lock, async.out, async.input)
	return async
}

func (async *Async) run(as *sync.WaitGroup) {
	defer as.Done()
	// defer fmt.Println("End run")
	// c := 0
	outchan := async.out
	lock := sync.WaitGroup{}
	tick := time.NewTicker(5 * time.Second)
	Finish := 0
	// canEND := false
	// O:
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
				fmt.Printf("\n[running : %d Finish: %d ]                 \r", c, Finish)
			default:
				if c < async.num*4 {
					break E
				}
			}
			time.Sleep(100 * time.Microsecond)

		}
		nsess := async.session.Copy()
		lock.Add(1)
		async.runningState <- 1
		go func(cond *sync.WaitGroup, out chan AsyncOut, running chan int, sess *Session, url string) {
			defer cond.Done()
			// fmt.Println("async:", url, sess.Proxy)

			with := sess.With(url)

			res := AsyncOut{
				Url: url,
				// Err: with.Err
				Res: with,
			}
			out <- res
			<-running
			Finish++
		}(&lock, outchan, async.runningState, nsess, inUrl)

	}
	lock.Wait()
	async.out <- AsyncOut{End: true}
}
