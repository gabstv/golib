package gstr

import (
	"strings"
)

func Split3(s, sep string) (a, b, c string) {
	s0 := strings.Split(s, sep)
	if len(s0) >= 3 {
		c = s0[2]
		b = s0[1]
		a = s0[0]
		return
	} else if len(s0) >= 2 {
		b = s0[1]
		a = s0[0]
		return
	} else if len(s0) >= 1 {
		a = s0[0]
		return
	}
	return
}

func Split2(s, sep string) (a, b string) {
	s0 := strings.Split(s, sep)
	if len(s0) >= 2 {
		b = s0[1]
		a = s0[0]
		return
	} else if len(s0) >= 1 {
		a = s0[0]
		return
	}
	return
}

func LenGtAll(limit int, rest ...string) bool {
	for _, v := range rest {
		if len(v) <= limit {
			return false
		}
	}
	return true
}
