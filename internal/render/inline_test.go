package render

import (
	"strings"
	"testing"

	"github.com/amterp/go-delta/internal/align"
	"github.com/amterp/go-delta/internal/diff"
)

// markerStyles wraps text with markers so we can assert styling in tests.
func markerStyles() Styles {
	return Styles{
		Removed:     func(s string) string { return "[R:" + s + "]" },
		Added:       func(s string) string { return "[A:" + s + "]" },
		RemovedEmph: func(s string) string { return "[RE:" + s + "]" },
		AddedEmph:   func(s string) string { return "[AE:" + s + "]" },
		LineNum:     func(s string) string { return "[N:" + s + "]" },
		Separator:   func(s string) string { return "[S:" + s + "]" },
		Plain:       func(s string) string { return s },
	}
}

func TestRenderInlineEmpty(t *testing.T) {
	result := RenderInline(nil, NoColorStyles())
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestRenderInlineUnpairedLines(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: "hello"},
					{Kind: diff.OpInsert, Content: "world"},
				},
			},
			// No pairs - completely different lines
		},
	}
	s := markerStyles()
	result := RenderInline(hunks, s)
	// Unpaired: entire line styled with base color
	if !strings.Contains(result, "[R:- hello]") {
		t.Errorf("expected unpaired removed line, got:\n%s", result)
	}
	if !strings.Contains(result, "[A:+ world]") {
		t.Errorf("expected unpaired added line, got:\n%s", result)
	}
}

func TestRenderInlinePairedLinesHaveEmphasis(t *testing.T) {
	// Create a hunk where lines are paired with alignment data
	oldTokens := align.Tokenize(`"age": 30,`)
	newTokens := align.Tokenize(`"age": 31,`)
	alignment := align.Align(oldTokens, newTokens)

	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: `"age": 30,`},
					{Kind: diff.OpInsert, Content: `"age": 31,`},
				},
			},
			Pairs: []align.LinePair{
				{OldIdx: 0, NewIdx: 1, Alignment: alignment},
			},
		},
	}
	s := markerStyles()
	result := RenderInline(hunks, s)

	// Paired lines should have emphasis markers on changed tokens
	if !strings.Contains(result, "[RE:") {
		t.Errorf("expected removed emphasis marker, got:\n%s", result)
	}
	if !strings.Contains(result, "[AE:") {
		t.Errorf("expected added emphasis marker, got:\n%s", result)
	}
	// The changed token "30" should be in removed emphasis
	if !strings.Contains(result, "[RE:30]") {
		t.Errorf("expected '30' in removed emphasis, got:\n%s", result)
	}
	// The changed token "31" should be in added emphasis
	if !strings.Contains(result, "[AE:31]") {
		t.Errorf("expected '31' in added emphasis, got:\n%s", result)
	}
}

func TestRenderInlineContextLines(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpEqual, Content: "context"},
					{Kind: diff.OpDelete, Content: "old"},
					{Kind: diff.OpInsert, Content: "new"},
					{Kind: diff.OpEqual, Content: "context2"},
				},
			},
		},
	}
	s := NoColorStyles()
	result := RenderInline(hunks, s)
	if !strings.Contains(result, "context") {
		t.Errorf("expected context line, got:\n%s", result)
	}
	// Context lines have space prefix
	if !strings.Contains(result, "  context") {
		t.Errorf("expected context line with space prefix, got:\n%s", result)
	}
}

func TestRenderInlineHunkSeparator(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1, Skipped: 0,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: "a"},
				},
			},
		},
		{
			Hunk: diff.Hunk{
				OldStart: 10, NewStart: 10, Skipped: 5,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: "b"},
				},
			},
		},
	}
	s := markerStyles()
	result := RenderInline(hunks, s)
	if !strings.Contains(result, "5 lines skipped") {
		t.Errorf("expected '5 lines skipped', got:\n%s", result)
	}
}
