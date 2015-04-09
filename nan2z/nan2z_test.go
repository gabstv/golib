package nan2z

import (
	"math"
	"testing"
)

type tdm struct {
	Valid   float64
	Invalid float64
}

func Test0(t *testing.T) {
	zzz := &tdm{77, math.NaN()}
	var x, y, z float64
	x = 10
	y = 20
	z = math.NaN()
	in0 := struct {
		Valid   float64
		Invalid float64
		DifMap  map[string]interface{}
		TDM     *tdm
	}{5, math.NaN(), map[string]interface{}{"a": x, "b": y, "c": z}, zzz}
	hasnan, ok := Run(&in0)
	t.Log("OK:", ok)
	t.Log("HASNAN:", hasnan)
	if in0.Invalid != 0 {
		t.Fatal("in0.Invalid should be 0", in0)
	}
	if in0.TDM.Invalid != 0 {
		t.Fatal("in0.TDM.Invalid should be 0", in0)
	}
}
