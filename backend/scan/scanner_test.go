package scan

import "testing"

func TestParseLoraWeights(t *testing.T) {
	w := parseLoraWeights("[\"0.8\",\"0.7\"]")
	if len(w) != 2 || w[0] != 0.8 || w[1] != 0.7 {
		t.Fatalf("unexpected weights %v", w)
	}
}

func TestParseLoraWeightsComma(t *testing.T) {
	w := parseLoraWeights("0.8,0.7")
	if len(w) != 2 || w[0] != 0.8 || w[1] != 0.7 {
		t.Fatalf("unexpected weights %v", w)
	}
}
