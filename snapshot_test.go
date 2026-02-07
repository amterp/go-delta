package godelta

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateSnapshots = flag.Bool("update", false, "update snapshot files")

// snapshotTest compares a diff result against a snapshot file. If -update
// is passed, it writes the snapshot file instead.
func snapshotTest(t *testing.T, name string, result string) {
	t.Helper()
	path := filepath.Join("testdata", name+".snap")

	if *updateSnapshots {
		err := os.MkdirAll("testdata", 0o755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte(result), 0o644)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("snapshot file %s not found (run with -update to create): %v", path, err)
	}
	if string(expected) != result {
		t.Errorf("output does not match snapshot %s\n--- expected ---\n%s\n--- got ---\n%s",
			path, string(expected), result)
	}
}

// --- Snapshot tests ---

func TestSnapshotInlineBasic(t *testing.T) {
	old := `{"name": "Alice", "age": 30, "city": "NYC"}`
	new := `{"name": "Alice", "age": 31, "city": "NYC"}`
	result := DiffWith(old, new, WithColor(false))
	snapshotTest(t, "inline_basic", result)
}

func TestSnapshotInlineMultiLine(t *testing.T) {
	old := "line1\nline2\nline3\nline4\nline5"
	new := "line1\nLINE2\nline3\nline4\nLINE5"
	result := DiffWith(old, new, WithColor(false))
	snapshotTest(t, "inline_multiline", result)
}

func TestSnapshotInlineMultiHunk(t *testing.T) {
	old := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl"
	new := "a\nB\nc\nd\ne\nf\ng\nh\ni\nj\nK\nl"
	result := DiffWith(old, new, WithColor(false), WithContextLines(1))
	snapshotTest(t, "inline_multihunk", result)
}

func TestSnapshotInlineContextZero(t *testing.T) {
	old := "a\nb\nc\nd\ne"
	new := "a\nB\nc\nD\ne"
	result := DiffWith(old, new, WithColor(false), WithContextLines(0))
	snapshotTest(t, "inline_context0", result)
}

func TestSnapshotInlineContextFive(t *testing.T) {
	old := "a\nb\nc\nd\ne\nf\ng\nh"
	new := "a\nb\nc\nD\ne\nf\ng\nh"
	result := DiffWith(old, new, WithColor(false), WithContextLines(5))
	snapshotTest(t, "inline_context5", result)
}

func TestSnapshotInlineAllAdded(t *testing.T) {
	result := DiffWith("", "alpha\nbeta\ngamma", WithColor(false))
	snapshotTest(t, "inline_all_added", result)
}

func TestSnapshotInlineAllRemoved(t *testing.T) {
	result := DiffWith("alpha\nbeta\ngamma", "", WithColor(false))
	snapshotTest(t, "inline_all_removed", result)
}

func TestSnapshotInlineTrailingNewline(t *testing.T) {
	result := DiffWith("hello\nworld\n", "hello\nworld", WithColor(false))
	snapshotTest(t, "inline_trailing_newline", result)
}

func TestSnapshotInlineEmphasis(t *testing.T) {
	old := `  "name": "Alice",`
	new := `  "name": "Bob",`
	result := DiffWith(old, new, WithColor(false))
	snapshotTest(t, "inline_emphasis", result)
}

func TestSnapshotInlineUnpaired(t *testing.T) {
	old := "completely different"
	new := "!@#$%^&*()"
	result := DiffWith(old, new, WithColor(false))
	snapshotTest(t, "inline_unpaired", result)
}

func TestSnapshotSideBySideBasic(t *testing.T) {
	old := "hello world\nfoo bar"
	new := "hello earth\nfoo baz"
	result := DiffWith(old, new, WithColor(false), WithSideBySide(true), WithWidth(80))
	snapshotTest(t, "sbs_basic", result)
}

func TestSnapshotSideBySideUnmatched(t *testing.T) {
	old := "only in old\nshared"
	new := "shared\nonly in new"
	result := DiffWith(old, new, WithColor(false), WithSideBySide(true), WithWidth(80))
	snapshotTest(t, "sbs_unmatched", result)
}

func TestSnapshotSideBySideLongLine(t *testing.T) {
	old := "short"
	new := "this is a very long line that should exceed the panel width in an 80-column terminal and get truncated"
	result := DiffWith(old, new, WithColor(false), WithSideBySide(true), WithWidth(80))
	snapshotTest(t, "sbs_long_line", result)
}

func TestSnapshotSideBySideWideUnicode(t *testing.T) {
	old := "hello 世界"
	new := "hello 地球"
	result := DiffWith(old, new, WithColor(false), WithSideBySide(true), WithWidth(80))
	snapshotTest(t, "sbs_wide_unicode", result)
}
