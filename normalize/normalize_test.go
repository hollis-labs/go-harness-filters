package normalize

import "testing"

// stubNormalizer is the minimum contract impl needed to verify the
// interface compiles. Concrete normalizers (slug, singularize, alias)
// land in follow-up files.
type stubNormalizer struct{}

func (stubNormalizer) Normalize(input string) Result {
	return Result{Original: input, Normalized: input}
}

func TestNormalizerContract(t *testing.T) {
	var n Normalizer = stubNormalizer{}
	r := n.Normalize("Hello")
	if r.Original != "Hello" {
		t.Errorf("Original = %q, want Hello", r.Original)
	}
	if r.Normalized != "Hello" {
		t.Errorf("Normalized = %q, want Hello", r.Normalized)
	}
}
