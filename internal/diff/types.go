package diff

// OpKind classifies a line in a diff.
type OpKind int

const (
	OpEqual  OpKind = iota // line is unchanged
	OpDelete               // line exists only in the old text
	OpInsert               // line exists only in the new text
)

// Line represents a single line in a diff result.
type Line struct {
	Kind    OpKind
	Content string // the line text (without trailing newline)
}

// Hunk is a contiguous group of diff lines with surrounding context.
type Hunk struct {
	OldStart int    // 1-based line number in the old text
	NewStart int    // 1-based line number in the new text
	Lines    []Line // the lines in this hunk (context + changes)
	Skipped  int    // number of lines skipped before this hunk
}
