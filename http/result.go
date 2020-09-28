package http

import (
	"encoding/json"
	"fmt"
	"time"
)

type Resulter struct {
	Oks  []string
	Errs []error
}

func (r *Resulter) PutErr(err error) {
	r.Errs = append(r.Errs, fmt.Errorf("%s | %v", time.Now(), err))
}

func (r *Resulter) PutOk(m string) {
	r.Oks = append(r.Oks, time.Now().String()+" | "+m)
}

func (r *Resulter) Json() string {
	o, _ := json.MarshalIndent(r, "", "\t")
	return string(o)
}

// func (r *Resulter) Html() string {

// }
