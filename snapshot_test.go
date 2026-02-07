package godelta

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
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

// ansiToMarkers replaces ANSI SGR sequences with readable markers.
// e.g. \x1b[31m -> «31», \x1b[0m -> «0», \x1b[32;7m -> «32;7»
var sgrRe = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

func ansiToMarkers(s string) string {
	return sgrRe.ReplaceAllString(s, `«$1»`)
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
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutSideBySide), WithWidth(80))
	snapshotTest(t, "sbs_basic", result)
}

func TestSnapshotSideBySideUnmatched(t *testing.T) {
	old := "only in old\nshared"
	new := "shared\nonly in new"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutSideBySide), WithWidth(80))
	snapshotTest(t, "sbs_unmatched", result)
}

func TestSnapshotSideBySideLongLine(t *testing.T) {
	old := "short"
	new := "this is a very long line that should exceed the panel width in an 80-column terminal and get truncated"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutSideBySide), WithWidth(80))
	snapshotTest(t, "sbs_long_line", result)
}

func TestSnapshotSideBySideWideUnicode(t *testing.T) {
	old := "hello 世界"
	new := "hello 地球"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutSideBySide), WithWidth(80))
	snapshotTest(t, "sbs_wide_unicode", result)
}

func TestSnapshotSideBySideColorBasic(t *testing.T) {
	old := "hello world\nfoo bar"
	new := "hello earth\nfoo baz"
	result := DiffWith(old, new, WithColor(true), WithLayout(LayoutSideBySide), WithWidth(80))
	snapshotTest(t, "sbs_color_basic", ansiToMarkers(result))
}

func TestSnapshotSideBySideColorTruncation(t *testing.T) {
	// Reproduce the exact bug from the demo: emphasis on a paired line
	// gets truncated, and the reset that would close the emphasis is lost.
	old := `{
  "name": "Alice",
  "hobbies": ["reading", "hiking"],
  "address": {
    "city": "Springfield",
  }
}`
	new := `{
  "name": "Bob",
  "hobbies": ["reading", "cycling", "cooking", "gaming"],
  "address": {
    "city": "Springfield",
  }
}`
	result := DiffWith(old, new, WithColor(true), WithLayout(LayoutSideBySide), WithWidth(72))
	marked := ansiToMarkers(result)

	// Every line should have its ANSI state properly closed.
	// Specifically, the truncated hobbies line must end with a reset «0»
	// before the newline, not bleed into subsequent lines.
	snapshotTest(t, "sbs_color_truncation", marked)
}

func TestSnapshotPreferSBSFits(t *testing.T) {
	old := "hello world"
	new := "hello earth"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutPreferSideBySide), WithWidth(200))
	snapshotTest(t, "prefer_sbs_fits", result)
}

func TestSnapshotPreferSBSDoesNotFit(t *testing.T) {
	old := "hello world with some extra content here"
	new := "hello earth with some extra content here"
	result := DiffWith(old, new, WithColor(false), WithLayout(LayoutPreferSideBySide), WithWidth(40))
	snapshotTest(t, "prefer_sbs_no_fit", result)
}

func TestSnapshotInlineCaretShift(t *testing.T) {
	old := `  --> <script>:2:25
  |
1 |
2 | for idx, item, extra in [1, 2, 3]:
  |                         ^^^^^^^^^
3 |     print(idx, item, extra)
  |
   = help: The for-loop syntax changed. Use: for item in items with loop: print(loop.idx, item). See: https://amterp.github.io/rad/migrations/v0.7/
   = info: rad explain RAD20033`
	new := `error[RAD20033]: Cannot unpack "int" into 3 values
  --> <script>:2:25
  |
1 |
2 | for idx, item, extra in [1, 2, 3]:
  |                          ^^^^^^^
3 |     print(idx, item, extra)
  |
   = help: The for-loop syntax changed. Use: for item in items with loop: print(loop.idx, item). See: https://amterp.github.io/rad/migrations/v0.7/
   = info: rad explain RAD20033`
	result := DiffWith(old, new, WithColor(true))
	snapshotTest(t, "inline_caret_shift", ansiToMarkers(result))
}
