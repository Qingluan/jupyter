package http

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConsoleBar struct {
	All       int64
	Now       int64
	Width     int
	NowWidth  int
	Interval  int
	Last      time.Time
	LastMsg   string
	LastWrite string
	LastBar   string
}

var (
	locker = sync.Mutex{}
)

func (pro *ConsoleBar) Add(i int) int {
	pro.Now += int64(i)

	if time.Now().Sub(pro.Last) > time.Duration(pro.Interval)*time.Second {
		nowWidth := int(float32(pro.Now) / float32(pro.All) * float32(pro.Width))
		pro.NowWidth = nowWidth
		fmt.Print(pro.LastMsg, "\r")
		pro.Last = time.Now()
		pro.Write()
	}
	return int(pro.Now)
}

func (pro *ConsoleBar) Update() {
	// return pro.Add(1)
}

func (pro *ConsoleBar) Increment() int {
	return pro.Add(1)
}

func (pro *ConsoleBar) Reset() {
	pro.Now = 0
	pro.LastMsg = ""
}

func NewConsoleBar(all int64) (bar *ConsoleBar, err error) {
	bar = &ConsoleBar{
		All:      all,
		Interval: 1,
		Last:     time.Now(),
	}
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		c, err := cmd.Output()
		if err != nil {
			bar.Width = 50
			// log.Fatal(err)
		}
		size := strings.Fields(string(c))[1]
		o, ierr := strconv.Atoi(size)
		if ierr != nil {
			return nil, ierr
		}
		bar.Width = o - 2
	} else {
		bar.Width = 50
	}
	return
}

func (pro *ConsoleBar) Write(args ...interface{}) {
	var realMsg string
	if len(args) == 0 {
		if strings.HasPrefix(pro.LastMsg, ":") {
			realMsg = fmt.Sprintf("%.2f%% || %d/%d :", pro.GetPercent(), pro.Now, pro.All) + strings.SplitN(pro.LastMsg, ":", 2)[1]
		} else {
			realMsg = fmt.Sprintf("%.2f%% || %d/%d :", pro.GetPercent(), pro.Now, pro.All) + pro.LastMsg
		}
	} else {
		realMsg = fmt.Sprintf("%.2f%% || %d/%d :", pro.GetPercent(), pro.Now, pro.All) + Blue(args...)
		pro.LastMsg = realMsg
	}
	if len(realMsg) < pro.Width {
		realMsg += LANGEMPTY
	}
	realMsg = realMsg[:pro.Width]
	if len(realMsg) >= pro.NowWidth {
		pre := realMsg[:pro.NowWidth]
		pro.LastBar = BlueBack(pre) + realMsg[pro.NowWidth:] + "\r"
		fmt.Print(BlueBack(pre) + realMsg[pro.NowWidth:] + "\r")
	} else {
		realMsg += strings.Repeat(" ", pro.NowWidth-len(realMsg))
		//  = BlueBack(pre) + realMsg[pro.NowWidth:] + "\r"
		pre := realMsg[:pro.NowWidth]
		pro.LastBar = BlueBack(pre) + realMsg[pro.NowWidth:] + "\r"
		fmt.Print(BlueBack(pre) + realMsg[pro.NowWidth:] + "\r")

	}
}

func (pro *ConsoleBar) GetPercent() float32 {
	return float32(pro.Now) / float32(pro.All) * float32(100)
}

func (pro *ConsoleBar) Error(err error) {
	locker.Lock()
	defer locker.Unlock()
	fmt.Print(LANGEMPTY[:pro.Width] + "\r")
	fmt.Println(Red(fmt.Sprintf("[!] :")), Bold(err))
	fmt.Print(pro.LastBar)
}

func (pro *ConsoleBar) Println(args ...interface{}) {
	locker.Lock()
	defer locker.Unlock()
	fmt.Print(LANGEMPTY[:pro.Width] + "\r")
	fmt.Println(Green(fmt.Sprintf("[+] :")), strings.TrimSpace(Bold(args...)))
	fmt.Print(pro.LastBar)
}

func (pro *ConsoleBar) Finished() {
	locker.Lock()
	defer locker.Unlock()
	pro.Reset()
	fmt.Print(LANGEMPTY[:pro.Width] + "\r")
}

func (pro *ConsoleBar) SetAll(all int) {
	pro.All = int64(all)
}

func (pro *ConsoleBar) SetMsg(msg string) {
	pro.Write(msg)
}
