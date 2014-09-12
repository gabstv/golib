package mq

import (
	"encoding/json"
	"testing"
)

func Test0(t *testing.T) {
	jq := `{"bourbon":true,"peas":3}`
	vv := &QMap{}
	json.Unmarshal([]byte(jq), vv)
	if !vv.C("bourbon").Bool() {
		t.Errorf("%v != %v", `vv.C("bourbon").Bool()`, true)
	}
	if vv.C("peas").Int() != 3 {
		t.Errorf("%v != %v", `vv.C("peas").Int()`, 3)
	}
}
