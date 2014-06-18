package header

import (
	"bytes"
	"strings"
)

func MapParams(val string) map[string]string {
	vs := strings.SplitN(val, ";", 2)
	outp := make(map[string]string)
	outp["value"] = strings.TrimSpace(vs[0])
	if len(vs) < 2 {
		return outp
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
					outp[strings.TrimSpace(lastk.String())] = strings.TrimSpace(lastv.String())
					lastk.Truncate(0)
					lastv.Truncate(0)
					insideval = false
					insidekey = true
					insidequot = false
				}
			}
		}
	}
	if lastk.Len() > 0 {
		k := strings.TrimSpace(lastk.String())
		outp[k] = strings.TrimSpace(lastv.String())
	}
	return outp
}
