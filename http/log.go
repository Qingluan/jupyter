package http

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	LANGEMPTY = "                                                                                                                                                                                                                                                                                                                                                                                                       "
)

var (
	Red       = color.New(color.FgRed).SprintFunc()
	Green     = color.New(color.FgGreen).SprintFunc()
	GreenBack = color.New(color.BgGreen, color.Bold).SprintFunc()
	BlueBack  = color.New(color.BgBlue, color.FgHiWhite, color.Bold).SprintFunc()
	Yello     = color.New(color.FgYellow).SprintFunc()
	Blue      = color.New(color.FgBlue).SprintFunc()
	Magenta   = color.New(color.FgMagenta).SprintFunc()
	Bold      = color.New(color.Bold).SprintFunc()
	Underline = color.New(color.Underline).SprintFunc()
	Hblue     = color.New(color.FgHiBlue).SprintFunc()
	Hgreen    = color.New(color.FgHiGreen).SprintFunc()
	Hyello    = color.New(color.FgHiYellow).SprintFunc()
)

func Success(args ...interface{}) {
	nargs := append([]interface{}{Green("[+]")}, Bold(args...))
	fmt.Println(nargs...)
}

func Failed(args ...interface{}) {
	nargs := append([]interface{}{Red("[-]")}, Underline(args...))
	fmt.Println(nargs...)
}

func Info(args ...interface{}) {
	nargs := append([]interface{}{Blue("[?]")}, args...)
	fmt.Println(nargs...)
}

func ProgressLog(now int, all int, msg string) {
	nargs := append([]interface{}{Blue(fmt.Sprintf("[%d/%d]", now, all))}, msg)
	nargs = append(nargs, "\r")
	fmt.Print(nargs...)
}

func L(args ...interface{}) {
	nargs := append([]interface{}{Blue("[-]")}, args...)
	nargs = append(nargs, "\r")
	fmt.Print(nargs...)

}

type Loger interface {
	Println(args ...interface{})
	Error(error)
	SetMsg(msg string)
}

type defaultloger struct{}

func (log *defaultloger) Println(args ...interface{}) {
	// fmt.Print()
	fmt.Println("\r"+LANGEMPTY[:30]+"\r"+Green(fmt.Sprintf("[+] :")), strings.TrimSpace(Bold(args...)))
}

func (log *defaultloger) Error(err error) {
	// fmt.Print()
	fmt.Println("\r"+LANGEMPTY[:30]+"\r"+Red(fmt.Sprintf("[!] :")), Bold(err.Error()))
}

func (log *defaultloger) SetMsg(msg string) {
	fmt.Printf("%s\r", strings.TrimSpace(msg))
}
