package normalize

import "testing"

// stubNormalizer is the minimum contract impl needed to verify the
// interface compiles.
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

func TestSlugNormalizerCanonicalizes(t *testing.T) {
	n := NewSlugNormalizer(Config{})
	got := n.Normalize(" Front End/API.v2 ")
	if got.Rejected {
		t.Fatalf("Normalize rejected: %+v", got)
	}
	if got.Normalized != "front-end-api-v2" {
		t.Errorf("Normalized = %q", got.Normalized)
	}
	if got.Original != " Front End/API.v2 " {
		t.Errorf("Original = %q", got.Original)
	}
}

func TestSlugNormalizerAliases(t *testing.T) {
	n := NewSlugNormalizer(Config{
		Aliases: map[string]string{"front end": "frontend"},
	})
	got := n.Normalize("Front End")
	if got.Normalized != "frontend" {
		t.Errorf("Normalized = %q", got.Normalized)
	}
	if got.AliasOf != "frontend" {
		t.Errorf("AliasOf = %q", got.AliasOf)
	}
	if got.Reason != "normalize.alias" {
		t.Errorf("Reason = %q", got.Reason)
	}
}

func TestSlugNormalizerRejectsReserved(t *testing.T) {
	n := NewSlugNormalizer(Config{Reserved: []string{"admin"}})
	got := n.Normalize("Admin")
	if !got.Rejected {
		t.Fatal("reserved term was not rejected")
	}
	if got.Reason != "normalize.reserved" {
		t.Errorf("Reason = %q", got.Reason)
	}
}

func TestSlugNormalizerSingularizesSimpleWords(t *testing.T) {
	n := SlugNormalizer{Singularize: true}
	got := n.Normalize("knowledge bases")
	if got.Normalized != "knowledge-base" {
		t.Errorf("Normalized = %q", got.Normalized)
	}
}

func TestSlugNormalizerRejectsEmpty(t *testing.T) {
	n := NewSlugNormalizer(Config{})
	got := n.Normalize(" ! ")
	if !got.Rejected {
		t.Fatal("empty normalized value was not rejected")
	}
	if got.Reason != "normalize.empty" {
		t.Errorf("Reason = %q", got.Reason)
	}
}
