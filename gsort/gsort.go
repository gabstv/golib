package gsort

import (
	"reflect"
	"regexp"
	"sort"
	"strconv"
)

var (
	re0 = regexp.MustCompile(`^([0-9]+\.?[0-9]*)`)
)

type ss []string

func (v ss) Len() int {
	return len(v)
}

func (v ss) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v ss) Less(i, j int) bool {
	return v[i] < v[j]
}

type sn []string

func (v sn) Len() int {
	return len(v)
}

func (v sn) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v sn) Less(i, j int) bool {
	match_i := re0.FindStringSubmatch(v[i])
	match_j := re0.FindStringSubmatch(v[j])
	if len(match_i) != len(match_j) {
		return v[i] < v[j]
	}
	if len(match_i) < 2 {
		return v[i] < v[j]
	}
	f0, _ := strconv.ParseFloat(match_i[1], 64)
	f1, _ := strconv.ParseFloat(match_j[1], 64)
	return f0 < f1
}

func fv(v interface{}) reflect.Value {
	v0 := reflect.ValueOf(v)
	for v0.Kind() == reflect.Ptr {
		v0 = v0.Elem()
	}
	return v0
}

func SortedStringKeys(v interface{}, isnumeric bool) []string {
	v0 := fv(v)
	if v0.Kind() != reflect.Map {
		return nil
	}
	keys := v0.MapKeys()
	if len(keys) == 0 {
		return []string{}
	}
	if keys[0].Kind() != reflect.String {
		return nil
	}
	out0 := make([]string, len(keys))
	for a, b := range keys {
		out0[a] = b.String()
	}
	// now sort
	if isnumeric {
		sort.Sort(sn(out0))
	}
	sort.Sort(ss(out0))
	return out0
}
