package errbuffer

import (
	"bytes"
	"errors"
	"log"
	"strconv"
)

type errb struct {
	errors []error
	titles []string
}

type errbf0 struct {
	parent *errb
}

func New() *errb {
	v := &errb{[]error{}, []string{}}
	return v
}

func (e *errb) Add2(err error, title string) {
	if err == nil {
		return
	}
	e.errors = append(e.errors, err)
	e.titles = append(e.titles, title)
}

func (e *errb) Add(err error) *errbf0 {
	if err == nil {
		return &errbf0{}
	}
	if len(e.errors) > len(e.titles) {
		e.titles = append(e.titles, "")
	}
	e.errors = append(e.errors, err)
	return &errbf0{e}
}

func (e *errbf0) Title(v string) {
	if e.parent == nil {
		return
	}
	if len(e.parent.errors) <= len(e.parent.titles) {
		return
	}
	e.parent.titles = append(e.parent.titles, v)
}

func (e *errb) Combined() error {
	if len(e.errors) < 1 {
		return nil
	}
	buffer := new(bytes.Buffer)
	for k := range e.errors {
		buffer.WriteString(strconv.Itoa(k + 1))
		if len(e.titles) > k {
			if len(e.titles[k]) > 0 {
				buffer.WriteString(" (" + e.titles[k] + ")")
			}
		}
		buffer.WriteString(": ")
		buffer.WriteString(e.errors[k].Error())
		buffer.WriteString("\n")
	}
	return errors.New(buffer.String())
}

func (e *errb) Log(errorTitle, successTitle string) {
	if len(e.errors) < 1 {
		log.Println(successTitle)
		return
	}
	er9 := e.Combined()
	log.Println(errorTitle, er9)
}
