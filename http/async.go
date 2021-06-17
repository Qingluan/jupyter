package http

import (
	"fmt"
	"log"
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
	num     int
	running int
	lock    sync.WaitGroup
	input   chan string
	out     chan AsyncOut
	session *Session
	do      func(out *AsyncOut)
}

func (session *Session) StartAsync(i int) *Async {
	async := new(Async)
	async.input = make(chan string, i)
	async.out = make(chan AsyncOut, i)
	async.num = i
	async.session = session.Copy()
	async.lock.Add(1)
	go async.run(&async.lock)

	return async
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
	log.Println("Wait to End:", async.running)
	async.lock.Wait()
	return async.session
}

func (async *Async) Each(do func(out *AsyncOut)) *Async {
	async.do = do
	async.lock.Add(1)
	go func(cal func(res *AsyncOut), l *sync.WaitGroup, out chan AsyncOut) {
		defer l.Done()
		defer fmt.Println("End Each")
		for {
			outres := <-out
			// log.Println("Out:", outres.Url)
			if outres.End {
				break
			}
			cal(&outres)
			// *cal--

		}
	}(async.do, &async.lock, async.out)
	return async
}

func (async *Async) run(as *sync.WaitGroup) {
	defer as.Done()
	defer fmt.Println("End run")
	// c := 0
	outchan := async.out
	lock := sync.WaitGroup{}
	for {
		inUrl := <-async.input
		// async.lock.Add(1)
		if inUrl == "[END]" {
			break
		}
		// log.Println("async:", inUrl)
		// c++
		nsess := async.session.Copy()
		lock.Add(1)
		go func(cond *sync.WaitGroup, out chan AsyncOut, sess *Session, url string) {
			defer cond.Done()
			// defer counter.Done()
			// sess.With(url)
			if sess.Proxy != "" {
				sess.SetSocks5Proxy(sess.Proxy)
			}
			with := sess.With(url)
			// fmt.Println("async:", with.URL, sess.Proxy)

			res := AsyncOut{
				Url: url,
				// Err: with.Err
				Res: with,
			}
			out <- res

		}(&lock, outchan, nsess, inUrl)

	}
	lock.Wait()
	async.out <- AsyncOut{End: true}
}
