package http

import (
	"fmt"
	"strconv"
)

type Value struct {
	v interface{}
}

type Dict map[string]string

type NextValue func(v Value) Value
type Gfunc map[string]NextValue

func (v Value) Increase() Value {
	return v.Add(1)
}

func NewValue(i interface{}) Value {
	return Value{
		v: i,
	}
}

func (v Value) Empty() bool {
	if v.v == nil {
		return true
	}
	return false
}

func (v Value) Add(one int) Value {
	switch v.v.(type) {
	case int:
		a := v.v.(int) + one
		v.v = a
	case string:
		s, err := strconv.Atoi(v.v.(string))
		if err == nil {
			v.v = fmt.Sprintf("%d", s+one)
		} else {
			Failed("can not conver to int")
		}
	case float32:
		v.v = v.v.(float32) + float32(one)
	case float64:
		v.v = v.v.(float64) + float64(one)
	}
	return v
}

func (v Value) AsInt() (int, error) {
	return strconv.Atoi(v.v.(string))
}

func (v Value) String() string {
	return fmt.Sprintf("%v", v.v)
}
