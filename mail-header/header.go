package header

import (
	"bytes"
	"log"
	"regexp"
	"strings"
)

var (
	Verbose = false
)

// Returns a lowercade bound map of strings
func MapParams(val string) map[string]string {
	if Verbose {
		log.Println("MapParams", val)
	}
	vs := strings.SplitN(val, ";", 2)
	outp := make(map[string]string)
	outp["value"] = strings.TrimSpace(vs[0])
	if len(vs) < 2 {
		if ok, _ := regexp.Match("^\\w+=", []byte(outp["value"])); ok {
			vs = append(vs, outp["value"])
		} else {
			return outp
		}
	} else {
		if ok, _ := regexp.Match("^\\w+=", []byte(outp["value"])); ok {
			vs[1] = vs[1] + "; " + vs[0]
		}
	}
	lastk := new(bytes.Buffer)
	lastv := new(bytes.Buffer)
	var insidekey, insidequot, insideval bool
	insidekey = true
	for k, v := range vs[1] {
		if insidekey {
			if v == '=' {
				insidekey = false
				insideval = true
				lastv.Truncate(0)
				continue
			}
			lastk.WriteRune(v)
		} else if insideval {
			if v == '"' {
				if insidequot && k > 0 {
					if vs[1][k-1] == '\\' {
						lastv.WriteRune('"')
					} else {
						insidequot = false
					}
				} else if !insidequot {
					if lastv.Len() < 2 {
						insidequot = true
					}
				}
			} else if v == ';' {
				if insidequot {
					lastv.WriteRune(';')
				} else {
					if Verbose {
						log.Println(lastk.String(), "=", lastv.String())
					}
					outp[strings.ToLower(strings.TrimSpace(lastk.String()))] = strings.TrimSpace(lastv.String())
					lastk.Truncate(0)
					lastv.Truncate(0)
					insideval = false
					insidekey = true
					insidequot = false
				}
			} else {
				lastv.WriteRune(v)
			}
		}
	}
	if lastk.Len() > 0 {
		k := strings.ToLower(strings.TrimSpace(lastk.String()))
		outp[k] = strings.TrimSpace(lastv.String())
	}
	return outp
}

func ExtractNameEmail(fromstr string) (string, string) {
	f2 := strings.Split(fromstr, "<")
	if len(f2) != 2 {
		nm := strings.TrimSpace(fromstr)
		nm = strings.Trim(nm, "\"")
		return nm, nm
	}
	f2[1] = strings.Trim(f2[1], "> ")
	f2[0] = strings.TrimSpace(f2[0])
	f2[0] = strings.Trim(f2[0], "\"")
	return f2[0], f2[1]
}
