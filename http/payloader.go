package http

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Payloader string
type G map[string]interface{}

var (
	KeyEx = regexp.MustCompile(`\{\w+\}`)
)

// AsFile as file to open
func (pay Payloader) AsFile(howHandle func(f *os.File, err error)) {
	f, err := os.Open(string(pay))
	if err == nil {
		defer f.Close()
	}
	howHandle(f, err)
}

func (pay Payloader) String() string {
	return string(pay)
}

// Format by {}
func (pay Payloader) Format(args ...interface{}) Payloader {
	num := 0
	raw := string(pay)
	for strings.Contains(raw, "{}") {
		raw = strings.Replace(raw, "{}", fmt.Sprint(args[num]), 1)
	}
	return Payloader(raw)
}

func (pay Payloader) Lines() (a ArrayFilter) {
	for _, v := range strings.Split(pay.String(), "\n") {

		v = strings.TrimSpace(v)
		if v != "" {
			a = a.Add(v)
		}
	}
	return a
}

// Format by map {key}
func (pay Payloader) FormatMap(args map[string]interface{}) Payloader {
	// num := 0
	raw := string(pay)
	for _, k := range KeyEx.FindAllString(pay.String(), -1) {
		if v, ok := args[k[1:len(k)-1]]; ok {
			raw = strings.ReplaceAll(raw, k, fmt.Sprint(v))
		}
	}
	return Payloader(raw)
}
