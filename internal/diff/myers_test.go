package diff

import (
	"testing"
)

func TestDiffIdentical(t *testing.T) {
	result := Diff("hello\nworld", "hello\nworld")
	if result != nil {
		t.Errorf("identical strings should produce nil, got %v", result)
	}
}

func TestDiffBothEmpty(t *testing.T) {
	result := Diff("", "")
	if result != nil {
		t.Errorf("two empty strings should produce nil, got %v", result)
	}
}

func TestDiffOldEmpty(t *testing.T) {
	result := Diff("", "hello\nworld")
	expected := []Line{
		{Kind: OpInsert, Content: "hello"},
		{Kind: OpInsert, Content: "world"},
	}
	assertLines(t, expected, result)
}

func TestDiffNewEmpty(t *testing.T) {
	result := Diff("hello\nworld", "")
	expected := []Line{
		{Kind: OpDelete, Content: "hello"},
		{Kind: OpDelete, Content: "world"},
	}
	assertLines(t, expected, result)
}

func TestDiffSingleLineChange(t *testing.T) {
	result := Diff("hello", "world")
	expected := []Line{
		{Kind: OpDelete, Content: "hello"},
		{Kind: OpInsert, Content: "world"},
	}
	assertLines(t, expected, result)
}

func TestDiffMultiLineWithContext(t *testing.T) {
	old := "a\nb\nc\nd\ne"
	new := "a\nb\nX\nd\ne"
	result := Diff(old, new)
	expected := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpEqual, Content: "b"},
		{Kind: OpDelete, Content: "c"},
		{Kind: OpInsert, Content: "X"},
		{Kind: OpEqual, Content: "d"},
		{Kind: OpEqual, Content: "e"},
	}
	assertLines(t, expected, result)
}

func TestDiffInsertionInMiddle(t *testing.T) {
	old := "a\nc"
	new := "a\nb\nc"
	result := Diff(old, new)
	expected := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpInsert, Content: "b"},
		{Kind: OpEqual, Content: "c"},
	}
	assertLines(t, expected, result)
}

func TestDiffDeletionInMiddle(t *testing.T) {
	old := "a\nb\nc"
	new := "a\nc"
	result := Diff(old, new)
	expected := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpDelete, Content: "b"},
		{Kind: OpEqual, Content: "c"},
	}
	assertLines(t, expected, result)
}

func TestDiffTrailingNewlineMismatch(t *testing.T) {
	// "hello\n" splits to ["hello", ""], "hello" splits to ["hello"]
	old := "hello\n"
	new := "hello"
	result := Diff(old, new)
	// The trailing newline creates an extra empty-string line
	expected := []Line{
		{Kind: OpEqual, Content: "hello"},
		{Kind: OpDelete, Content: ""},
	}
	assertLines(t, expected, result)
}

func TestDiffMultipleChanges(t *testing.T) {
	old := "a\nb\nc\nd\ne\nf"
	new := "a\nB\nc\nd\nE\nf"
	result := Diff(old, new)
	expected := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpDelete, Content: "b"},
		{Kind: OpInsert, Content: "B"},
		{Kind: OpEqual, Content: "c"},
		{Kind: OpEqual, Content: "d"},
		{Kind: OpDelete, Content: "e"},
		{Kind: OpInsert, Content: "E"},
		{Kind: OpEqual, Content: "f"},
	}
	assertLines(t, expected, result)
}

func TestDiffAllAdded(t *testing.T) {
	result := Diff("", "a\nb\nc")
	expected := []Line{
		{Kind: OpInsert, Content: "a"},
		{Kind: OpInsert, Content: "b"},
		{Kind: OpInsert, Content: "c"},
	}
	assertLines(t, expected, result)
}

func TestDiffAllRemoved(t *testing.T) {
	result := Diff("a\nb\nc", "")
	expected := []Line{
		{Kind: OpDelete, Content: "a"},
		{Kind: OpDelete, Content: "b"},
		{Kind: OpDelete, Content: "c"},
	}
	assertLines(t, expected, result)
}

func assertLines(t *testing.T, expected, got []Line) {
	t.Helper()
	if len(expected) != len(got) {
		t.Errorf("length mismatch: expected %d lines, got %d", len(expected), len(got))
		t.Errorf("expected: %v", expected)
		t.Errorf("got:      %v", got)
		return
	}
	for i := range expected {
		if expected[i] != got[i] {
			t.Errorf("line %d: expected %v, got %v", i, expected[i], got[i])
		}
	}
}
