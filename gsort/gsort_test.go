package gsort

import (
	"testing"
)

func TestSort(t *testing.T) {
	undi := []string{"30 und", "10 und", "15 und", "5 und"}
	undf := []string{"5 und", "10 und", "15 und", "30 und"}
	undi = SortedStringKeys(undi, true)
	for k := range undi {
		if undi[k] != undf[k] {
			t.Fatal("undi", undi[k], "undf", undf[k])
		}
	}
}

func TestSort2(t *testing.T) {
	undia := map[string]float64{"30 und": 300, "10 und": 100, "15 und": 200, "5 und": 50}
	undf := []string{"5 und", "10 und", "15 und", "30 und"}
	undi := SortedStringKeys(undia, true)
	for k := range undi {
		if undi[k] != undf[k] {
			t.Fatal("undi", undi[k], "undf", undf[k])
		}
	}
}
