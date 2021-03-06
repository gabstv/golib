package header

import (
	"testing"
)

func TestMap(t *testing.T) {
	Verbose = true
	expect := map[string]string{"value": "image/jpeg", "name": "123.jpg", "filename": "12345.png"}
	raw := "image/jpeg; name=\"123.jpg\"; Filename=12345.png"
	result := MapParams(raw)
	for k, v := range expect {
		if v != result[k] {
			t.Errorf("expected `%v` got `%v`\n", v, result[k])
		}
	}
}

func TestMap2(t *testing.T) {
	Verbose = true
	expect := map[string]string{"value": "boundary=\"potato\"", "boundary": "potato"}
	raw := "boundary=\"potato\""
	result := MapParams(raw)
	for k, v := range expect {
		if v != result[k] {
			t.Errorf("expected `%v` got `%v`\n", v, result[k])
		}
	}
}

func TestMap3(t *testing.T) {
	Verbose = true
	expect := map[string]string{"value": "boundary=\"_004_5857d86ac1e145538e299fb48014a39bGRUPR80MB362lamprd80pro_\"", "boundary": "_004_5857d86ac1e145538e299fb48014a39bGRUPR80MB362lamprd80pro_", "type": "multipart/alternative"}
	raw := "boundary=\"_004_5857d86ac1e145538e299fb48014a39bGRUPR80MB362lamprd80pro_\"; type=\"multipart/alternative\""
	result := MapParams(raw)
	for k, v := range expect {
		if v != result[k] {
			t.Errorf("expected `%v` got `%v`\n", v, result[k])
		}
	}
}

func TestExtractMil(t *testing.T) {
	en := "Gabriel Ochsenhofer"
	ev := "gabriel@nutripele.com"
	raw := "Gabriel Ochsenhofer <gabriel@nutripele.com>"
	r0, r1 := ExtractNameEmail(raw)
	if r0 != en {
		t.Errorf("expected `%v` got `%v`\n", en, r0)
	}
	if r1 != ev {
		t.Errorf("expected `%v` got `%v`\n", ev, r1)
	}
	raw = "gabriel@nutripele.com"
	r0, r1 = ExtractNameEmail(raw)
	if r0 != "gabriel@nutripele.com" {
		t.Errorf("expected `%v` got `%v`\n", "gabriel@nutripele.com", r0)
	}
	if r1 != "gabriel@nutripele.com" {
		t.Errorf("expected `%v` got `%v`\n", "gabriel@nutripele.com", r1)
	}
}
