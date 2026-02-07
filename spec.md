# go-delta Specification

## Overview

go-delta is a Go library for rendering human-readable, colored diffs of text in the terminal. Its primary use case is snapshot testing: you have an expected string and a got string, and when they don't match, you want clear, colored output showing exactly what changed - down to the word level within changed lines.

It is a library, not a CLI tool. Import it and call a function. No shelling out.

### Why This Exists

Go's testing ecosystem lacks a good library for rendering pretty diffs. Existing Go diff libraries compute diffs but don't render them well. Delta (Rust) solved this for git, but nothing like it exists as an importable Go library. go-delta fills that gap.

## Module and Package

- **Module path**: `github.com/amterp/go-delta`
- **Package name**: `godelta`
- **Idiomatic import alias**: `gd`
- **Usage**: `gd.Diff(expected, got)`

## Requirements

### Core API (v0.1)

```go
package godelta

// Diff computes and renders a colored diff between two strings.
// Returns a string with ANSI escape codes for terminal display.
// Returns empty string if inputs are identical.
func Diff(expected, got string) string

// DiffWith computes a diff using the provided options.
func DiffWith(expected, got string, opts ...Option) string
```

### Options (v0.1)

```go
func WithContextLines(n int) Option     // lines of context around changes (default: 3)
func WithSideBySide(on bool) Option     // side-by-side layout (default: false, inline)
func WithColor(on bool) Option          // force color on/off (default: auto-detect TTY)
```

### Deferred to Post-v0.1

These are explicitly out of scope for the first version:

- `DiffBytes(expected, got []byte) string`
- `WithLineNumbers(on bool) Option`
- `WithHeader(expected, got string) Option`
- `WithWordDiff(on bool) Option`
- `ExpectMatch(t *testing.T, expected, got string)`

They may be added later but are not part of this specification.

## Behavior

### Identical Inputs

`Diff(a, a)` returns an empty string `""`.

### Empty Inputs

- `Diff("", "hello\nworld")` - everything shows as added (green).
- `Diff("hello\nworld", "")` - everything shows as removed (red).
- `Diff("", "")` - returns empty string.

### Trailing Newlines

A trailing newline mismatch (one string ends with `\n`, the other doesn't) is treated as a regular character difference. No special `\ No newline at end of file` marker.

### Color Behavior

- **TTY detected**: Full ANSI color output.
- **No TTY** (piped, CI): Strip ANSI codes. The `+`/`-` prefix characters remain - they are structural, not decorative.
- **`NO_COLOR` env var**: Respect the [no-color.org](https://no-color.org) convention. Same as no-TTY behavior.
- **`FORCE_COLOR` env var**: Force color output regardless of TTY detection.
- **`WithColor(true/false)`**: Overrides all auto-detection.

The `+`/`-` line prefixes are always present regardless of color mode.

## Diff Pipeline

The diff is computed and rendered in three stages.

### Stage 1: Line-Level Diff (Myers Algorithm)

Compute a standard line-level diff to identify added, removed, and unchanged lines. Group changes into hunks with configurable context lines (default: 3).

**Implementation**: Vendor the Myers diff algorithm from `golang.org/x/tools/internal/diff/myers` (~251 lines, BSD licensed). This is what gopls used, it's clean Go, and it's the most battle-tested Go diff implementation available.

The vendored code goes in `internal/diff/`. Adapt the output to produce line-level operations (add, remove, equal) rather than byte-offset edits.

### Stage 2: Within-Line Highlighting

For each pair of removed+added lines in a hunk, run a secondary word-level diff to identify which specific segments changed within the line. This is the key feature that separates useful diff output from noise.

#### Line Pairing

Follow Delta's greedy forward-search strategy:

1. For each removed line in a hunk, scan forward through the remaining added lines.
2. Tokenize both lines on word boundaries (`\w+` regex) with operators and punctuation as single-character tokens. For example, `foo.bar(baz)` tokenizes to `["foo", ".", "bar", "(", "baz", ")"]`.
3. Run a Needleman-Wunsch alignment on the token sequences.
4. Compute a normalized edit distance: `(deleted + inserted) / (2 * matched + deleted + inserted)`.
5. If the distance is below a threshold (hardcoded, ~0.6), pair the lines. Otherwise, leave them unpaired.
6. Move forward - order is preserved, no backtracking.

Unpaired lines display as full-line additions or removals without within-line emphasis.

**Rationale**: This is Delta's approach, adapted for Go. The greedy strategy is simple, preserves line ordering, and produces good results. The distance threshold prevents misleading highlights when lines are completely rewritten. The 1:1 greedy approach was chosen over more complex best-matching algorithms because it handles the common case well and is straightforward to implement and debug.

#### Within-Line Diff

For paired lines, the Needleman-Wunsch alignment from the pairing step already identifies which tokens are deletions, insertions, or unchanged (no-ops). Use these annotations directly to determine which segments to emphasize.

**Rationale**: Reusing the alignment from the pairing step avoids computing the diff twice. Word-level tokenization (vs. character-level) produces more readable output for the target use case (JSON, YAML, config files, template output).

#### Visual Treatment

Changed segments within a line are rendered with **ANSI reverse video** (attribute `\x1b[7m`), which swaps the foreground and background colors. On a removed line (red text), the changed segment becomes red background with default foreground. On an added line (green text), the changed segment becomes green background with default foreground.

This is the approach used by both Delta and git's diff-highlight. It works on any terminal that supports basic ANSI attributes (effectively all modern terminals), doesn't require 256-color support, and provides strong visual contrast between "this line changed" and "this specific part changed."

### Stage 3: Rendering

Assemble the final output string.

#### Color Scheme

| Element | Style |
|---------|-------|
| Removed lines | Red foreground |
| Added lines | Green foreground |
| Changed segments (within-line) | Reverse video on top of red/green |
| Context lines | Default terminal color (no dimming) |
| Line numbers | Dimmed / faint |
| Hunk separators | Dimmed / faint |
| `+`/`-`/` ` prefixes | Same color as the line they belong to |

#### Inline Mode (Default)

Standard unified-style layout. Removed lines above, added lines below, with context lines around them.

```
  12 │   "name": "Alice",
  13 │ - "age": 30,
  14 │ + "age": 31,
  15 │   "city": "NYC"
```

Where `30` on the removed line and `31` on the added line have reverse video applied (shown here in brackets for illustration only):

```
  13 │ - "age": [30],     ← 30 has reverse video (red bg)
  14 │ + "age": [31],     ← 31 has reverse video (green bg)
```

**Line numbers**: Shown in the gutter, dimmed. Each side has its own line number that tracks independently (removed lines increment the left number, added lines increment the right number, context lines increment both). In inline mode, show both numbers side by side in the gutter.

**Hunk separators**: When context lines are omitted between hunks, show a separator line indicating how many lines were skipped. Format: `~~~ N lines skipped ~~~` (dimmed). This makes it clear that content exists between the visible hunks and communicates the gap size.

**Blank line**: Insert a visual blank line between hunks for readability.

#### Side-by-Side Mode

Left panel shows the old text, right panel shows the new text. The two panels are separated by a `│` character.

- **Colors**: Left panel uses red for changed/removed lines, right panel uses green for changed/added lines. Within-line reverse video emphasis applies to both sides. Context lines appear on both sides in default color.
- **Line numbers**: Each panel has its own line number column.
- **Terminal width**: Detect terminal width using `golang.org/x/term.GetSize()`. Split available columns evenly between the two panels (minus gutter). If terminal width can't be detected, assume 80 columns.
- **Narrow terminals**: Do not silently fall back to inline mode. The user explicitly asked for side-by-side, so honor it even if lines wrap. Lines that exceed panel width wrap within the panel.
- **Removed-only lines**: Left panel shows the line in red, right panel is empty (or shows a placeholder like `~` dimmed).
- **Added-only lines**: Left panel is empty (or `~` dimmed), right panel shows the line in green.

#### Line Prefix Characters

Every line has a prefix character:

- ` ` (space) for context lines
- `-` for removed lines
- `+` for added lines

These are always present regardless of color mode. In no-color mode, they are the only way to distinguish line types.

## Dependencies

### External Dependencies

| Dependency | Purpose | Notes |
|------------|---------|-------|
| `github.com/amterp/color` | Terminal color output | ANSI color codes, bold, dim, reverse |
| `golang.org/x/term` | TTY detection, terminal width | `IsTerminal()`, `GetSize()`. Quasi-stdlib. |

### Vendored Code

| Source | What | License | Location |
|--------|------|---------|----------|
| `golang.org/x/tools/internal/diff/myers` | Myers diff algorithm | BSD 3-Clause | `internal/diff/` |

The vendored Myers code is adapted: stripped to essentials, output format changed to produce line-level operations suitable for rendering. Include the BSD license notice in the vendored files.

The within-line alignment (Needleman-Wunsch on tokens) is written from scratch, informed by Delta's `src/edits.rs` and `src/align.rs`.

## Non-Goals

- Syntax-aware / AST-aware diffing
- Git integration or unified diff parsing
- Binary file diffing
- File-level diffing (comparing directory trees)
- Syntax highlighting of diff content

## Design Decisions

### Myers over LCS for v0.1

The golang/tools `internal/diff` package offers both a 251-line Myers implementation and a newer ~780-line LCS-based approach. Myers is simpler, well-understood, and sufficient for the target use case (test output of tens to hundreds of lines). The LCS approach is more efficient for very large inputs with common prefixes/suffixes, but performance is not critical for this library. Start with Myers; switch to LCS later if performance becomes an issue.

### Greedy Line Pairing over Best-Match

Delta's greedy forward-search pairing was chosen over optimal matching (e.g., Hungarian algorithm) because:
- It preserves line ordering, which matches human expectations.
- It's simple to implement and debug.
- It handles the common case (small, localized edits) well.
- Optimal matching is O(n^3) and only marginally better for typical diffs.

### Word-Level over Character-Level Tokenization

Word-level tokenization (with single-char operators) was chosen as the default because:
- For the target content (JSON, YAML, config files), word boundaries are the natural unit of change.
- Character-level diffs produce noisy highlights on word substitutions (e.g., changing `alice` to `bob` highlights every character individually).
- Word-level groups changes into meaningful chunks.

### Reverse Video for Within-Line Emphasis

Reverse video was chosen over bold, underline, or background-only because:
- It works on any terminal with basic ANSI support.
- It provides strong contrast between "line changed" and "this specific part changed."
- It's what Delta and git diff-highlight both use.
- It doesn't require 256-color or truecolor support.

### Always-Present Prefix Characters

`+`/`-`/` ` prefixes are always emitted, even in color mode, because:
- They provide structural information that's useful even with color.
- They make the output parseable in no-color mode.
- They're familiar from unified diff format.
- They make it easy to pipe output or copy-paste without losing meaning.

## Open Questions

None. All requirements are specified to implementation-ready detail.
