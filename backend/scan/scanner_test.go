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

func TestExtractPromptLoras(t *testing.T) {
        res := extractPromptLoras("a <lora:foo:0.5> and <lyco:bar:1>")
        if len(res) != 2 {
                t.Fatalf("expected 2 loras, got %d", len(res))
        }
        if res[0].name != "foo" || res[0].weight == nil || *res[0].weight != 0.5 {
                t.Fatalf("unexpected first lora %+v", res[0])
        }
        if res[1].name != "bar" || res[1].weight == nil || *res[1].weight != 1 {
                t.Fatalf("unexpected second lora %+v", res[1])
        }
}
