package align

import (
	"math"
	"testing"
)

func TestAlignBothEmpty(t *testing.T) {
	a := Align(nil, nil)
	if len(a.Old) != 0 || len(a.New) != 0 {
		t.Errorf("both empty should produce empty alignment")
	}
	if a.Distance != 0 {
		t.Errorf("expected distance 0, got %f", a.Distance)
	}
}

func TestAlignOldEmpty(t *testing.T) {
	tokens := []Token{{Text: "a"}, {Text: "b"}}
	a := Align(nil, tokens)
	if len(a.New) != 2 {
		t.Fatalf("expected 2 new tokens, got %d", len(a.New))
	}
	for _, at := range a.New {
		if at.Op != AlignInsert {
			t.Errorf("expected insert, got %d", at.Op)
		}
	}
	if a.Distance != 1.0 {
		t.Errorf("expected distance 1.0, got %f", a.Distance)
	}
}

func TestAlignNewEmpty(t *testing.T) {
	tokens := []Token{{Text: "a"}, {Text: "b"}}
	a := Align(tokens, nil)
	if len(a.Old) != 2 {
		t.Fatalf("expected 2 old tokens, got %d", len(a.Old))
	}
	for _, at := range a.Old {
		if at.Op != AlignDelete {
			t.Errorf("expected delete, got %d", at.Op)
		}
	}
}

func TestAlignIdentical(t *testing.T) {
	old := Tokenize("hello world")
	new := Tokenize("hello world")
	a := Align(old, new)
	if a.Distance != 0 {
		t.Errorf("identical should have distance 0, got %f", a.Distance)
	}
	for _, at := range a.Old {
		if at.Op != AlignMatch {
			t.Errorf("expected all matches in old, got %d", at.Op)
		}
	}
	for _, at := range a.New {
		if at.Op != AlignMatch {
			t.Errorf("expected all matches in new, got %d", at.Op)
		}
	}
}

func TestAlignSingleTokenChange(t *testing.T) {
	old := Tokenize(`"age": 30,`)
	new := Tokenize(`"age": 31,`)
	a := Align(old, new)

	// Distance should be low - only one token changed
	if a.Distance > 0.3 {
		t.Errorf("single token change should have low distance, got %f", a.Distance)
	}

	// Check that "30" -> "31" is the changed pair
	hasOldDelete := false
	for _, at := range a.Old {
		if at.Op == AlignDelete && at.Token.Text == "30" {
			hasOldDelete = true
		}
	}
	hasNewInsert := false
	for _, at := range a.New {
		if at.Op == AlignInsert && at.Token.Text == "31" {
			hasNewInsert = true
		}
	}
	if !hasOldDelete {
		t.Error("expected '30' to be marked as delete")
	}
	if !hasNewInsert {
		t.Error("expected '31' to be marked as insert")
	}
}

func TestAlignCompletelyDifferent(t *testing.T) {
	old := Tokenize("foo bar baz")
	new := Tokenize("xxx yyy zzz")
	a := Align(old, new)

	// All tokens changed, but structure (spaces) is the same.
	// Words are all different, spaces match.
	// Distance should be moderately high.
	if a.Distance < 0.3 {
		t.Errorf("completely different words should have higher distance, got %f", a.Distance)
	}
}

func TestAlignNormalizedDistance(t *testing.T) {
	// Two tokens: one matches, one changed
	old := []Token{{Text: "a"}, {Text: "b"}}
	new := []Token{{Text: "a"}, {Text: "c"}}
	a := Align(old, new)

	// 1 match, 1 delete, 1 insert
	// distance = (1+1) / (2*1 + 1 + 1) = 2/4 = 0.5
	if math.Abs(a.Distance-0.5) > 0.01 {
		t.Errorf("expected distance ~0.5, got %f", a.Distance)
	}
}
