package align

import (
	"strings"
	"testing"
)

func TestTokenizeEmpty(t *testing.T) {
	tokens := Tokenize("")
	if tokens != nil {
		t.Errorf("empty string should produce nil, got %v", tokens)
	}
}

func TestTokenizeSimpleWords(t *testing.T) {
	tokens := Tokenize("hello world")
	assertTokenTexts(t, tokens, []string{"hello", " ", "world"})
}

func TestTokenizePunctuation(t *testing.T) {
	tokens := Tokenize("foo.bar(baz)")
	assertTokenTexts(t, tokens, []string{"foo", ".", "bar", "(", "baz", ")"})
}

func TestTokenizeJSON(t *testing.T) {
	tokens := Tokenize(`"age": 30,`)
	assertTokenTexts(t, tokens, []string{`"`, "age", `"`, ":", " ", "30", ","})
}

func TestTokenizeWhitespace(t *testing.T) {
	tokens := Tokenize("  a  b  ")
	assertTokenTexts(t, tokens, []string{" ", " ", "a", " ", " ", "b", " ", " "})
}

func TestTokenizeUnderscore(t *testing.T) {
	tokens := Tokenize("foo_bar baz")
	assertTokenTexts(t, tokens, []string{"foo_bar", " ", "baz"})
}

func TestTokenizeDigits(t *testing.T) {
	tokens := Tokenize("x123+456")
	assertTokenTexts(t, tokens, []string{"x123", "+", "456"})
}

func TestTokenizeUnicode(t *testing.T) {
	tokens := Tokenize("héllo wörld")
	assertTokenTexts(t, tokens, []string{"héllo", " ", "wörld"})
}

func TestTokenizeLossless(t *testing.T) {
	inputs := []string{
		"hello world",
		"foo.bar(baz)",
		`"age": 30,`,
		"  a  b  ",
		"héllo wörld",
		"a+b*c/d",
		"",
	}
	for _, input := range inputs {
		tokens := Tokenize(input)
		var texts []string
		for _, tok := range tokens {
			texts = append(texts, tok.Text)
		}
		rejoined := strings.Join(texts, "")
		if rejoined != input {
			t.Errorf("lossless check failed: %q -> %q", input, rejoined)
		}
	}
}

func TestTokenizePositions(t *testing.T) {
	tokens := Tokenize("ab cd")
	// "ab" at [0,2), " " at [2,3), "cd" at [3,5)
	if tokens[0].Start != 0 || tokens[0].End != 2 {
		t.Errorf("token 0: expected [0,2), got [%d,%d)", tokens[0].Start, tokens[0].End)
	}
	if tokens[1].Start != 2 || tokens[1].End != 3 {
		t.Errorf("token 1: expected [2,3), got [%d,%d)", tokens[1].Start, tokens[1].End)
	}
	if tokens[2].Start != 3 || tokens[2].End != 5 {
		t.Errorf("token 2: expected [3,5), got [%d,%d)", tokens[2].Start, tokens[2].End)
	}
}

func assertTokenTexts(t *testing.T, tokens []Token, expected []string) {
	t.Helper()
	if len(tokens) != len(expected) {
		var got []string
		for _, tok := range tokens {
			got = append(got, tok.Text)
		}
		t.Fatalf("expected %d tokens %v, got %d tokens %v", len(expected), expected, len(tokens), got)
	}
	for i, tok := range tokens {
		if tok.Text != expected[i] {
			t.Errorf("token %d: expected %q, got %q", i, expected[i], tok.Text)
		}
	}
}
