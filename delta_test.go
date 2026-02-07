package godelta

import (
	"strings"
	"testing"
)

func TestDiffIdenticalReturnsEmpty(t *testing.T) {
	result := Diff("hello\nworld", "hello\nworld")
	if result != "" {
		t.Errorf("identical strings should return empty, got %q", result)
	}
}

func TestDiffBothEmptyReturnsEmpty(t *testing.T) {
	result := Diff("", "")
	if result != "" {
		t.Errorf("both empty should return empty, got %q", result)
	}
}

func TestDiffSimpleChange(t *testing.T) {
	result := DiffWith("a\nb\nc", "a\nB\nc", WithColor(false))
	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(result, "- b") {
		t.Errorf("expected removed line '- b', got:\n%s", result)
	}
	if !strings.Contains(result, "+ B") {
		t.Errorf("expected added line '+ B', got:\n%s", result)
	}
}

func TestDiffAllAdded(t *testing.T) {
	result := DiffWith("", "a\nb\nc", WithColor(false))
	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(result, "+ a") {
		t.Errorf("expected '+ a' in output, got:\n%s", result)
	}
}

func TestDiffAllRemoved(t *testing.T) {
	result := DiffWith("a\nb\nc", "", WithColor(false))
	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(result, "- a") {
		t.Errorf("expected '- a' in output, got:\n%s", result)
	}
}

func TestDiffContextLines(t *testing.T) {
	old := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
	new := "1\n2\n3\n4\nX\n6\n7\n8\n9\n10"
	result := DiffWith(old, new, WithColor(false), WithContextLines(1))
	if !strings.Contains(result, "lines skipped") {
		t.Errorf("expected 'lines skipped' separator, got:\n%s", result)
	}
}

func TestDiffContextZero(t *testing.T) {
	old := "a\nb\nc"
	new := "a\nB\nc"
	result := DiffWith(old, new, WithColor(false), WithContextLines(0))
	if strings.Contains(result, "  a") {
		t.Errorf("context 0 should not show context lines, got:\n%s", result)
	}
}

func TestDiffLineNumbers(t *testing.T) {
	result := DiffWith("a\nb", "a\nB", WithColor(false))
	if !strings.Contains(result, "2") {
		t.Errorf("expected line number 2, got:\n%s", result)
	}
}

func TestDiffHunkSeparator(t *testing.T) {
	var oldLines, newLines []string
	for i := 1; i <= 20; i++ {
		s := "x"
		if i == 2 {
			oldLines = append(oldLines, "old2")
			newLines = append(newLines, "new2")
		} else if i == 18 {
			oldLines = append(oldLines, "old18")
			newLines = append(newLines, "new18")
		} else {
			oldLines = append(oldLines, s)
			newLines = append(newLines, s)
		}
	}
	old := strings.Join(oldLines, "\n")
	new := strings.Join(newLines, "\n")
	result := DiffWith(old, new, WithColor(false), WithContextLines(2))
	if !strings.Contains(result, "lines skipped") {
		t.Errorf("expected hunk separator, got:\n%s", result)
	}
}

func TestDiffTrailingNewline(t *testing.T) {
	result := DiffWith("hello\n", "hello", WithColor(false))
	if result == "" {
		t.Fatal("trailing newline difference should produce a diff")
	}
}

func TestDiffSingleCharLines(t *testing.T) {
	result := DiffWith("a", "b", WithColor(false))
	if result == "" {
		t.Fatal("single char diff should produce output")
	}
	if !strings.Contains(result, "- a") || !strings.Contains(result, "+ b") {
		t.Errorf("expected proper single-char diff, got:\n%s", result)
	}
}

func TestDiffWhitespaceOnlyLines(t *testing.T) {
	result := DiffWith("  \n  ", "  \n    ", WithColor(false))
	if result == "" {
		t.Fatal("whitespace diff should produce output")
	}
}

func TestDiffVeryLongLine(t *testing.T) {
	long := strings.Repeat("x", 500)
	result := DiffWith(long, long+"y", WithColor(false))
	if result == "" {
		t.Fatal("long line diff should produce output")
	}
}

func TestDiffHunkOnlyAdds(t *testing.T) {
	result := DiffWith("a\nc", "a\nb\nc", WithColor(false))
	if !strings.Contains(result, "+ b") {
		t.Errorf("expected inserted line, got:\n%s", result)
	}
}

func TestDiffHunkOnlyRemoves(t *testing.T) {
	result := DiffWith("a\nb\nc", "a\nc", WithColor(false))
	if !strings.Contains(result, "- b") {
		t.Errorf("expected removed line, got:\n%s", result)
	}
}

func TestDiffSideBySideIntegration(t *testing.T) {
	old := "hello world\ngoodbye"
	new := "hello earth\ngoodbye"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutSideBySide), WithWidth(80))
	if result == "" {
		t.Fatal("side-by-side should produce output")
	}
	if !strings.Contains(result, " │ ") {
		t.Errorf("expected panel separator, got:\n%s", result)
	}
}

func TestDiffPreferSideBySideFitsUseSBS(t *testing.T) {
	old := "aaa\nbbb"
	new := "aaa\nccc"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutPreferSideBySide), WithWidth(200))
	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if !isSideBySide(result) {
		t.Errorf("wide terminal should use SBS, got:\n%s", result)
	}
}

func TestDiffPreferSideBySideNarrowFallsBackToInline(t *testing.T) {
	// Use multi-line content with a shared context line to ensure both
	// panels are wide enough that the SBS layout won't fit in 40 cols.
	old := "context: the quick brown fox jumps\nold line here"
	new := "context: the quick brown fox jumps\nnew line here"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutPreferSideBySide), WithWidth(40))
	if result == "" {
		t.Fatal("expected non-empty diff")
	}
	if isSideBySide(result) {
		t.Errorf("narrow terminal should fall back to inline, got:\n%s", result)
	}
}

// isSideBySide returns true if the output appears to be SBS format.
// SBS lines have 3 "│" characters per content line (left gutter,
// panel separator, right gutter), while inline has only 1.
func isSideBySide(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.Count(line, "│") >= 3 {
			return true
		}
	}
	return false
}

func TestDiffNegativeContextClampedToZero(t *testing.T) {
	result := DiffWith("a\nb\nc", "a\nB\nc", WithColor(false), WithContextLines(-5))
	if result == "" {
		t.Fatal("negative context should be treated as 0")
	}
}

func TestWithColorTrue(t *testing.T) {
	result := DiffWith("a", "b", WithColor(true))
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("WithColor(true) should produce ANSI escapes, got:\n%q", result)
	}
}

func TestWithColorFalse(t *testing.T) {
	result := DiffWith("a", "b", WithColor(false))
	if strings.Contains(result, "\x1b[") {
		t.Errorf("WithColor(false) should not produce ANSI escapes, got:\n%q", result)
	}
}

func TestDiffUnicode(t *testing.T) {
	result := DiffWith("héllo wörld", "héllo éarth", WithColor(false))
	if result == "" {
		t.Fatal("unicode diff should produce output")
	}
}
