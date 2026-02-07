package render

import (
	"strings"
	"testing"

	"github.com/amterp/go-delta/internal/align"
	"github.com/amterp/go-delta/internal/diff"
)

func TestRenderSideBySideEmpty(t *testing.T) {
	result := RenderSideBySide(nil, NoColorStyles(), 80)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestRenderSideBySideContextLines(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpEqual, Content: "same line"},
				},
			},
		},
	}
	result := RenderSideBySide(hunks, NoColorStyles(), 80)
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), result)
	}
	if !strings.Contains(lines[0], " │ ") {
		t.Errorf("expected panel separator, got: %q", lines[0])
	}
	parts := strings.SplitN(lines[0], " │ ", 3)
	if len(parts) < 3 {
		t.Fatalf("expected 3 parts split by separator, got %d", len(parts))
	}
}

func TestRenderSideBySideUnpairedRemove(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: "removed"},
				},
			},
		},
	}
	s := markerStyles()
	result := RenderSideBySide(hunks, s, 80)
	if !strings.Contains(result, "removed") {
		t.Errorf("expected 'removed' in output, got:\n%s", result)
	}
	if !strings.Contains(result, "[S:~]") {
		t.Errorf("expected dimmed '~' placeholder, got:\n%s", result)
	}
}

func TestRenderSideBySideUnpairedInsert(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpInsert, Content: "added"},
				},
			},
		},
	}
	s := markerStyles()
	result := RenderSideBySide(hunks, s, 80)
	if !strings.Contains(result, "added") {
		t.Errorf("expected 'added' in output, got:\n%s", result)
	}
	if !strings.Contains(result, "[S:~]") {
		t.Errorf("expected dimmed '~' on left, got:\n%s", result)
	}
}

func TestRenderSideBySidePairedLines(t *testing.T) {
	oldTokens := align.Tokenize("foo bar")
	newTokens := align.Tokenize("foo baz")
	alignment := align.Align(oldTokens, newTokens)

	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpDelete, Content: "foo bar"},
					{Kind: diff.OpInsert, Content: "foo baz"},
				},
			},
			Pairs: []align.LinePair{
				{OldIdx: 0, NewIdx: 1, Alignment: alignment},
			},
		},
	}
	s := markerStyles()
	result := RenderSideBySide(hunks, s, 80)
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (paired), got %d:\n%s", len(lines), result)
	}
	if !strings.Contains(result, "[RE:") {
		t.Errorf("expected removed emphasis, got:\n%s", result)
	}
	if !strings.Contains(result, "[AE:") {
		t.Errorf("expected added emphasis, got:\n%s", result)
	}
}

func TestRenderSideBySideHunkSeparator(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{Hunk: diff.Hunk{OldStart: 1, NewStart: 1, Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: "a"},
		}}},
		{Hunk: diff.Hunk{OldStart: 10, NewStart: 10, Skipped: 5, Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: "b"},
		}}},
	}
	s := markerStyles()
	result := RenderSideBySide(hunks, s, 80)
	if !strings.Contains(result, "5 lines skipped") {
		t.Errorf("expected '5 lines skipped', got:\n%s", result)
	}
}

func TestRenderSideBySideNarrowTerminal(t *testing.T) {
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpEqual, Content: "short"},
				},
			},
		},
	}
	result := RenderSideBySide(hunks, NoColorStyles(), 20)
	if result == "" {
		t.Error("narrow terminal should still produce output")
	}
}

func TestRenderSideBySideLongLineTruncated(t *testing.T) {
	longContent := strings.Repeat("x", 200)
	hunks := []align.AnnotatedHunk{
		{
			Hunk: diff.Hunk{
				OldStart: 1, NewStart: 1,
				Lines: []diff.Line{
					{Kind: diff.OpEqual, Content: longContent},
				},
			},
		},
	}
	result := RenderSideBySide(hunks, NoColorStyles(), 80)
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	for _, line := range lines {
		vw := visibleWidth(line)
		if vw > 80 {
			t.Errorf("line exceeds terminal width (%d > 80): %q", vw, line)
		}
	}
	// Truncated panels should end with "…"
	if !strings.Contains(result, "…") {
		t.Errorf("truncated content should show '…' indicator, got:\n%s", result)
	}
}

func TestTruncateToWidthAppendsResetWhenTruncatingANSI(t *testing.T) {
	// Simulate what happens in SBS: a styled line gets truncated.
	// The reset (\x1b[0m) that closes the style is in the truncated portion,
	// so truncateToWidth must append its own reset to prevent color bleed.
	styled := "\x1b[32;7mhello world\x1b[0m"
	result := truncateToWidth(styled, 5)

	// Must end with a reset sequence
	if !strings.HasSuffix(result, "\x1b[0m") {
		t.Errorf("truncated ANSI string should end with reset, got %q", result)
	}
	if visibleWidth(result) != 5 {
		t.Errorf("expected visible width 5, got %d for %q", visibleWidth(result), result)
	}
}

func TestTruncateToWidthNoResetWhenNotTruncated(t *testing.T) {
	// If the string fits, no reset should be appended
	styled := "\x1b[31mhi\x1b[0m"
	result := truncateToWidth(styled, 10)
	if result != styled {
		t.Errorf("non-truncated string should be unchanged, got %q", result)
	}
}

func TestTruncateToWidthNoResetWhenNoANSI(t *testing.T) {
	// Plain text truncation should not get a spurious reset
	result := truncateToWidth("hello world", 5)
	if strings.Contains(result, "\x1b") {
		t.Errorf("plain truncated string should not contain ANSI, got %q", result)
	}
}

// --- visibleWidth and truncateToWidth tests ---

func TestVisibleWidthASCII(t *testing.T) {
	if w := visibleWidth("hello"); w != 5 {
		t.Errorf("expected 5, got %d", w)
	}
}

func TestVisibleWidthANSI(t *testing.T) {
	// "\x1b[31m" is red, "\x1b[0m" is reset
	s := "\x1b[31mhello\x1b[0m"
	if w := visibleWidth(s); w != 5 {
		t.Errorf("expected 5 (ANSI stripped), got %d", w)
	}
}

func TestVisibleWidthWideUnicode(t *testing.T) {
	// CJK characters are 2 columns each
	s := "hello世界"
	if w := visibleWidth(s); w != 9 {
		t.Errorf("expected 9 (5 + 2*2), got %d", w)
	}
}

func TestVisibleWidthANSITerminators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"CSI with @ terminator", "\x1b[@hello", 5},
		{"CSI with ~ terminator", "\x1b[~hello", 5},
		{"CSI with letter terminator", "\x1b[31mhello", 5},
		{"bare ESC + digit (single-char escape)", "\x1b7hello", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := visibleWidth(tt.input); w != tt.want {
				t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, w, tt.want)
			}
		})
	}
}

func TestTruncateToWidthShort(t *testing.T) {
	result := truncateToWidth("hello", 10)
	if result != "hello" {
		t.Errorf("short string should be unchanged, got %q", result)
	}
}

func TestTruncateToWidthExact(t *testing.T) {
	result := truncateToWidth("hello", 5)
	if result != "hello" {
		t.Errorf("exact fit should be unchanged, got %q", result)
	}
}

func TestTruncateToWidthLong(t *testing.T) {
	result := truncateToWidth("hello world", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateToWidthPreservesANSI(t *testing.T) {
	s := "\x1b[31mhello world\x1b[0m"
	result := truncateToWidth(s, 5)
	// Should keep the opening escape and first 5 visible chars
	if !strings.Contains(result, "\x1b[31m") {
		t.Errorf("expected ANSI escape preserved, got %q", result)
	}
	if visibleWidth(result) != 5 {
		t.Errorf("expected visible width 5, got %d for %q", visibleWidth(result), result)
	}
}

func TestTruncateToWidthWideChar(t *testing.T) {
	// "ab世" = 4 columns. Truncate to 3 should give "ab " (space filler)
	result := truncateToWidth("ab世", 3)
	if visibleWidth(result) != 3 {
		t.Errorf("expected visible width 3, got %d for %q", visibleWidth(result), result)
	}
	if result != "ab " {
		t.Errorf("expected 'ab ' (space filler for partial wide char), got %q", result)
	}
}

// --- OSC/DCS/string-type escape tests ---

func TestVisibleWidthOSC(t *testing.T) {
	// OSC (window title): ESC ] 0 ; Title BEL
	s := "\x1b]0;Window Title\x07hello"
	if w := visibleWidth(s); w != 5 {
		t.Errorf("expected 5 (OSC skipped), got %d", w)
	}
}

func TestVisibleWidthOSCST(t *testing.T) {
	// OSC terminated by ST (ESC \) instead of BEL
	s := "\x1b]0;Window Title\x1b\\hello"
	if w := visibleWidth(s); w != 5 {
		t.Errorf("expected 5 (OSC+ST skipped), got %d", w)
	}
}

func TestVisibleWidthDCS(t *testing.T) {
	// DCS: ESC P ... ST
	s := "\x1bPsome data\x1b\\hello"
	if w := visibleWidth(s); w != 5 {
		t.Errorf("expected 5 (DCS skipped), got %d", w)
	}
}

func TestVisibleWidthAPC(t *testing.T) {
	// APC: ESC _ ... ST
	s := "\x1b_app command\x1b\\hello"
	if w := visibleWidth(s); w != 5 {
		t.Errorf("expected 5 (APC skipped), got %d", w)
	}
}

func TestTruncateToWidthPreservesOSC(t *testing.T) {
	s := "\x1b]0;Title\x07hello world"
	result := truncateToWidth(s, 5)
	if visibleWidth(result) != 5 {
		t.Errorf("expected visible width 5, got %d for %q", visibleWidth(result), result)
	}
	// The OSC sequence should be preserved in output
	if !strings.Contains(result, "\x1b]0;Title\x07") {
		t.Errorf("expected OSC preserved, got %q", result)
	}
}
