package http

import (
	"errors"
	"regexp"
	"sort"
)

type ArrayFilter []string
type DictBool map[string]bool

func (dict DictBool) Keys() (newArray ArrayFilter) {
	for k := range dict {
		newArray = append(newArray, k)
	}
	return
}

func (array ArrayFilter) In(o string) int {
	for i, k := range array {
		if k == o {
			return i
		}
	}
	return -1
}

func (array ArrayFilter) Add(o string) ArrayFilter {
	if array.In(o) < 0 {
		array = append(array, o)
	}
	return array
}
func (array ArrayFilter) Sort() ArrayFilter {
	sort.Strings(array)
	return array
}

func (array ArrayFilter) Filter(reStrOrFunc_str_bool interface{}) (newArray ArrayFilter) {
	switch reStrOrFunc_str_bool.(type) {
	case string:
		filterRe := regexp.MustCompile(reStrOrFunc_str_bool.(string))
		for _, v := range array {
			if len(filterRe.FindAllString(v, 1)) > 0 {
				newArray = append(newArray, v)
			}
		}
	// case func(no int,every string) bool:
	default:
		panic(errors.New("must func(int,string) bool / re str "))
	}
	return
}

func (array ArrayFilter) FilterFunc(reStrOrFunc_str_bool func(no int, every string) bool) (newArray ArrayFilter) {
	for i, v := range array {
		if reStrOrFunc_str_bool(i, v) {
			newArray = append(newArray, v)
		}
	}
	// case func(no int,every string) bool:
	return
}

func (array ArrayFilter) Every(handler func(no int, every string) string) ArrayFilter {
	for i, v := range array {
		array[i] = handler(i, v)
	}
	return array
}
