package http

import (
	"fmt"
	"time"
)

var (
	STOP         = "[STOP]"
	DefaultLoger = &defaultloger{}
)

type Result struct {
	Url string
	Res interface{}
}
type RunnerPool struct {
	Thread     int
	ready      []string
	errRecords map[string]int
	task       chan string
	result     chan Result
	doing      chan int
	Handle     func(arg string) interface{}
	After      func(res Result, loger Loger)
	ErrDo      func(error, int, Result, Loger)
	// Loger  Loger
	RetryTime int
	Bar       *ConsoleBar
}

func NewAwaitPool(thread int) (pool *RunnerPool) {
	pool = &RunnerPool{
		Thread:     thread,
		errRecords: make(map[string]int),
		task:       make(chan string),
		result:     make(chan Result),
		doing:      make(chan int, thread),
	}
	return pool
}

func (pool *RunnerPool) resultLoop() {

	for {
		res := <-pool.result
		if res.Url == STOP {
			break
		}

		switch res.Res.(type) {
		case error:
			if s, ok := pool.errRecords[res.Url]; ok {
				pool.errRecords[res.Url] = s + 1
			} else {
				pool.errRecords[res.Url] = 1
			}
			if pool.errRecords[res.Url] < pool.RetryTime {
				if pool.Bar != nil {
					pool.Bar.Error(fmt.Errorf("%s:%v", res.Url, res.Res))
				}
				pool.ready = append(pool.ready, res.Url)
				// continue
				if pool.ErrDo != nil {
					if pool.Bar != nil {
						pool.ErrDo(res.Res.(error), pool.errRecords[res.Url], res, pool.Bar)
					} else {
						pool.ErrDo(res.Res.(error), pool.errRecords[res.Url], res, DefaultLoger)
					}
					// continue
				}
			} else {

				if pool.Bar != nil {
					pool.Bar.Add(1)
				}
			}

		default:
			if pool.Bar != nil {
				pool.Bar.Add(1)
			}

		}

		if pool.After != nil {
			if pool.Bar != nil {
				pool.After(res, pool.Bar)
			} else {
				pool.After(res, DefaultLoger)
			}
		}
	}
}

func _run(handle func(arg string) interface{}, a string, doing chan int, res chan Result) {
	defer func() {
		<-doing
	}()
	doing <- 1
	res <- Result{
		Url: a,
		Res: handle(a),
	}
}
func (pool *RunnerPool) tasktLoop() {
	for {
		for {
			if len(pool.doing) < pool.Thread {
				break
			}
			time.Sleep(1 * time.Second)
		}
		arg := <-pool.task
		go _run(pool.Handle, arg, pool.doing, pool.result)
	}
}

func (pool *RunnerPool) Loop(args []string, showBar bool) {
	go pool.resultLoop()
	go pool.tasktLoop()
	go pool.Tick(3)
	if showBar {
		pool.Bar, _ = NewConsoleBar(int64(len(args)))
	}
	pool.ready = append(pool.ready, args...)

	for {
		if len(pool.ready) > 0 {
			var arg string
			for len(pool.ready) > 0 {
				arg, pool.ready = pool.ready[0], pool.ready[1:]
				pool.task <- arg
			}

			time.Sleep(2 * time.Second)
		}
		if len(pool.doing) == 0 {
			pool.task <- STOP
			break
		}
		time.Sleep(2 * time.Second)
	}
	return

}

func (pool *RunnerPool) Tick(sec int) {
	tick := time.Tick(time.Duration(sec) * time.Second)
	for {
		select {
		case <-tick:
			if pool.Bar != nil {
				// pool.Bar.SetMsg(fmt.Sprintf("Running:%d", len(pool.doing)))
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}

}