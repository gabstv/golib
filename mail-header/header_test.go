package header

import (
	"testing"
)

func TestMap(t *testing.T) {
	Verbose = true
	expect := map[string]string{"value": "image/jpeg", "name": "123.jpg", "filename": "12345.png"}
	raw := "image/jpeg; name=\"123.jpg\"; filename=12345.png"
	result := MapParams(raw)
	for k, v := range expect {
		if v != result[k] {
			t.Errorf("expected `%v` got `%v`\n", v, result[k])
		}
	}
}
