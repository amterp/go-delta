package godelta

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestInlineWidth(t *testing.T) {
	for _, tt := range []struct {
		name  string
		text  string
		width int
		parts []string
	}{
		{"words", "alpha beta gamma", 14, []string{"alpha ", "beta ", "gamma"}},
		{"exact fit", "abcdef", 14, []string{"abcdef"}},
		{"long word", "abcdefghijklm", 14, []string{"abcdef", "ghijkl", "m"}},
		{"whitespace", "  abc  def  ", 14, []string{"  abc ", " def  "}},
		{"unicode whitespace", "alpha\u2003beta", 14, []string{"alpha\u2003", "beta"}},
		{"wide characters", "世界地球", 13, []string{"世界", "地球"}},
		{"combining marks", "ae\u0301bc", 10, []string{"ae\u0301", "bc"}},
		{"spacing combining marks", "का", 9, []string{"का"}},
		{"mark on whitespace", "a \u0301bc", 10, []string{"a \u0301", "bc"}},
		{"tab stops", "\t123456789", 16, []string{"\t", "12345678", "9"}},
		{"tab after text", "a\t123456789", 16, []string{"a\t", "12345678", "9"}},
		{"narrow", "abc", 1, []string{"a", "b", "c"}},
		{"narrow wide characters", "世界", 9, []string{"世", "界"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffWith("", tt.text, WithColor(false), WithWidth(tt.width))
			var want strings.Builder
			for i, part := range tt.parts {
				if i == 0 {
					want.WriteString("  1 │ + ")
				} else {
					want.WriteString("        ")
				}
				want.WriteString(part)
				want.WriteByte('\n')
			}
			if got != want.String() {
				t.Fatalf("want %q, got %q", want.String(), got)
			}
		})
	}
}

func TestInlineWidthDisabled(t *testing.T) {
	old := "alpha beta gamma\nold value\nlast line"
	new := "alpha beta gamma\nnew value\nlast line"
	for _, color := range []bool{false, true} {
		want := DiffWith(old, new, WithColor(color))
		for _, width := range []int{-1, 0, 200} {
			got := DiffWith(old, new, WithColor(color), WithWidth(width))
			if got != want {
				t.Errorf("color %t, width %d changed unwrapped output\nwant %q\ngot  %q", color, width, want, got)
			}
		}
	}
}

func TestInlineWidthMultiDigitGutter(t *testing.T) {
	old := strings.Repeat("same\n", 10) + "removed words wrap"
	new := strings.Repeat("same\n", 10) + "added words wrap"
	got := DiffWith(old, new, WithColor(false), WithContextLines(0), WithWidth(16))
	if !strings.Contains(got, "11    │ - remove\n          d \n") ||
		!strings.Contains(got, "   11 │ + added \n          words \n          wrap\n") {
		t.Fatalf("wrong continuation indent or logical line number:\n%s", got)
	}
}

func TestInlineWidthEastAsianGutter(t *testing.T) {
	original := runewidth.DefaultCondition
	runewidth.DefaultCondition = &runewidth.Condition{EastAsianWidth: true, StrictEmojiNeutral: true}
	t.Cleanup(func() { runewidth.DefaultCondition = original })
	got := DiffWith("", "abc", WithColor(false), WithWidth(10))
	want := "  1 │ + a\n         b\n         c\n"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestInlineWidthEmbeddedANSI(t *testing.T) {
	for _, tt := range []struct {
		name string
		open string
		end  string
	}{
		{"SGR", "\x1b[34m", "\x1b[0m"},
		{"extended SGR", "\x1b[38;2;10;20;30m", "\x1b[0m"},
		{"combined SGR with zero channels", "\x1b[1;38;2;10;0;30m", "\x1b[0m"},
		{"indexed SGR", "\x1b[38;5;0m", "\x1b[0m"},
		{"indexed SGR padded mode", "\x1b[38;05;196m", "\x1b[0m"},
		{"colon SGR", "\x1b[38:2::10:0:30m", "\x1b[0m"},
		{"mixed delimiter SGR", "\x1b[38;2::255:0:0m", "\x1b[0m"},
		{"OSC hyperlink", "\x1b]8;;https://example.com\x1b\\", "\x1b]8;;\x1b\\"},
		{"OSC hyperlink BEL", "\x1b]8;;https://example.com\a", "\x1b]8;;\x1b\\"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.open + "alpha beta" + tt.end
			got := DiffWith("", text, WithColor(false), WithWidth(14))
			want := "  1 │ + " + tt.open + "alpha " + tt.end + "\n" +
				"        " + tt.open + "beta" + tt.end + "\n"
			if got != want {
				t.Fatalf("want %q, got %q", want, got)
			}
		})
	}
}

func TestInlineWidthSelectiveANSIResetsStayBounded(t *testing.T) {
	text := strings.Repeat("\x1b[31mx\x1b[39m", 128)
	got := DiffWith("", text, WithColor(false), WithWidth(9))
	if len(got) > 4*len(text) {
		t.Fatalf("ANSI replay grew output from %d to %d bytes", len(text), len(got))
	}
	for i, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		prefix := "        "
		if i == 0 {
			prefix = "  1 │ + "
		}
		if plain := sgrRe.ReplaceAllString(line, ""); plain != prefix+"x" {
			t.Fatalf("lost content or alignment on line %d: %q", i, plain)
		}
	}
}

func TestInlineWidthKeepsLogicalLines(t *testing.T) {
	got := DiffWith("", "alpha beta gamma\n\ndelta", WithColor(false), WithWidth(14))
	want := "  1 │ + alpha \n        beta \n        gamma\n  2 │ + \n  3 │ + delta\n"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestInlineWidthColorDoesNotChangeWrapping(t *testing.T) {
	old := "context words wrap\nset alpha beta gamma delta\nold only"
	new := "context words wrap\nset alpha zeta theta delta\n!@#$%^&*()"
	for width := 9; width <= 32; width++ {
		plain := DiffWith(old, new, WithColor(false), WithWidth(width))
		colored := DiffWith(old, new, WithColor(true), WithWidth(width))
		if got := sgrRe.ReplaceAllString(colored, ""); got != plain {
			t.Fatalf("width %d: color changed wrapping\nwant %q\ngot  %q", width, plain, got)
		}
	}
}

func TestInlineWidthPreservesContent(t *testing.T) {
	for _, text := range []string{
		"  alpha  beta gamma   ", "abcdefghijklm", "世界 地球", "ae\u0301b c", "  ",
	} {
		for width := 1; width <= 32; width++ {
			t.Run(fmt.Sprintf("%q/%d", text, width), func(t *testing.T) {
				got := DiffWith("", text, WithColor(false), WithWidth(width))
				var joined strings.Builder
				for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
					prefix := "        "
					if joined.Len() == 0 {
						prefix = "  1 │ + "
					}
					if !strings.HasPrefix(line, prefix) {
						t.Fatalf("wrong continuation indent: %q", line)
					}
					content := strings.TrimPrefix(line, prefix)
					joined.WriteString(content)
					limit := max(1, width-8)
					if columns := runewidth.StringWidth(content); columns > max(2, limit) {
						t.Fatalf("text exceeds available width %d: %q", limit, content)
					}
				}
				if joined.String() != text {
					t.Fatalf("lost content: want %q, got %q", text, joined.String())
				}
			})
		}
	}
}

func TestSnapshotInlineWrap(t *testing.T) {
	old := "context words span multiple rows\nset alpha beta gamma delta\nremoved text without a match\nlast context wraps too"
	new := "context words span multiple rows\nset alpha zeta theta delta\n!@#$%^&*()1234567890\nlast context wraps too"
	got := DiffWith(old, new, WithColor(false), WithWidth(24))
	snapshotTest(t, "inline_wrap", strings.ReplaceAll(got, "\n", "⏎\n"))
}

func TestSnapshotInlineWrapColor(t *testing.T) {
	old := "set abcdefghijklmnop value\nold only"
	new := "set abcdefuvwxyzmnop value\n!@#$%^&*()"
	got := DiffWith(old, new, WithColor(true), WithWidth(18))
	snapshotTest(t, "inline_wrap_color", ansiToMarkers(got))
}
