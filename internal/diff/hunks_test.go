package diff

import (
	"testing"
)

func TestComputeHunksEmpty(t *testing.T) {
	result := ComputeHunks(nil, 3)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestComputeHunksNoChanges(t *testing.T) {
	lines := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpEqual, Content: "b"},
	}
	result := ComputeHunks(lines, 3)
	if result != nil {
		t.Errorf("no changes should produce nil, got %v", result)
	}
}

func TestComputeHunksSingleChange(t *testing.T) {
	lines := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpEqual, Content: "b"},
		{Kind: OpDelete, Content: "c"},
		{Kind: OpInsert, Content: "C"},
		{Kind: OpEqual, Content: "d"},
		{Kind: OpEqual, Content: "e"},
	}
	hunks := ComputeHunks(lines, 3)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 1 || h.NewStart != 1 {
		t.Errorf("wrong start: old=%d, new=%d", h.OldStart, h.NewStart)
	}
	// context 3 around index 2-3, window is [0, 6) = all lines
	if len(h.Lines) != 6 {
		t.Errorf("expected 6 lines, got %d", len(h.Lines))
	}
	if h.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", h.Skipped)
	}
}

func TestComputeHunksContextZero(t *testing.T) {
	lines := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpEqual, Content: "b"},
		{Kind: OpDelete, Content: "c"},
		{Kind: OpInsert, Content: "C"},
		{Kind: OpEqual, Content: "d"},
		{Kind: OpEqual, Content: "e"},
	}
	hunks := ComputeHunks(lines, 0)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 3 || h.NewStart != 3 {
		t.Errorf("wrong start: old=%d, new=%d", h.OldStart, h.NewStart)
	}
	if len(h.Lines) != 2 {
		t.Errorf("expected 2 lines (delete+insert), got %d", len(h.Lines))
	}
	if h.Skipped != 2 {
		t.Errorf("expected 2 skipped, got %d", h.Skipped)
	}
}

func TestComputeHunksTwoSeparateChanges(t *testing.T) {
	// Two changes far apart should produce two hunks
	lines := []Line{
		{Kind: OpDelete, Content: "a"},   // 0
		{Kind: OpInsert, Content: "A"},   // 1
		{Kind: OpEqual, Content: "b"},    // 2
		{Kind: OpEqual, Content: "c"},    // 3
		{Kind: OpEqual, Content: "d"},    // 4
		{Kind: OpEqual, Content: "e"},    // 5
		{Kind: OpEqual, Content: "f"},    // 6
		{Kind: OpEqual, Content: "g"},    // 7
		{Kind: OpEqual, Content: "h"},    // 8
		{Kind: OpEqual, Content: "i"},    // 9
		{Kind: OpDelete, Content: "j"},   // 10
		{Kind: OpInsert, Content: "J"},   // 11
	}
	hunks := ComputeHunks(lines, 2)
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	// First hunk: change at 0-1, context 2 after = [0, 4)
	h0 := hunks[0]
	if h0.OldStart != 1 || h0.NewStart != 1 {
		t.Errorf("hunk 0 wrong start: old=%d, new=%d", h0.OldStart, h0.NewStart)
	}
	if len(h0.Lines) != 4 {
		t.Errorf("hunk 0 expected 4 lines, got %d", len(h0.Lines))
	}
	if h0.Skipped != 0 {
		t.Errorf("hunk 0 expected 0 skipped, got %d", h0.Skipped)
	}

	// Second hunk: change at 10-11, context 2 before = [8, 12)
	h1 := hunks[1]
	if len(h1.Lines) != 4 {
		t.Errorf("hunk 1 expected 4 lines, got %d", len(h1.Lines))
	}
	if h1.Skipped != 4 {
		t.Errorf("hunk 1 expected 4 skipped, got %d", h1.Skipped)
	}
}

func TestComputeHunksMergeOverlappingContext(t *testing.T) {
	// Two changes close enough that their context overlaps -> merged
	lines := []Line{
		{Kind: OpDelete, Content: "a"},
		{Kind: OpInsert, Content: "A"},
		{Kind: OpEqual, Content: "b"},
		{Kind: OpEqual, Content: "c"},
		{Kind: OpDelete, Content: "d"},
		{Kind: OpInsert, Content: "D"},
	}
	hunks := ComputeHunks(lines, 3)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 merged hunk, got %d", len(hunks))
	}
	if len(hunks[0].Lines) != 6 {
		t.Errorf("expected 6 lines, got %d", len(hunks[0].Lines))
	}
}

func TestComputeHunksLineNumbers(t *testing.T) {
	// Verify that OldStart and NewStart track correctly across hunks
	lines := Diff(
		"a\nb\nc\nd\ne\nf\ng\nh\ni\nj",
		"a\nB\nc\nd\ne\nf\ng\nh\ni\nJ",
	)
	hunks := ComputeHunks(lines, 1)
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	// First hunk around line 2 (b->B): old starts at 1, new starts at 1
	h0 := hunks[0]
	if h0.OldStart != 1 || h0.NewStart != 1 {
		t.Errorf("hunk 0: expected start (1,1), got (%d,%d)", h0.OldStart, h0.NewStart)
	}

	// Second hunk around line 10 (j->J)
	h1 := hunks[1]
	if h1.OldStart != 9 || h1.NewStart != 9 {
		t.Errorf("hunk 1: expected start (9,9), got (%d,%d)", h1.OldStart, h1.NewStart)
	}
}

func TestComputeHunksLargeContext(t *testing.T) {
	// Context larger than the total diff should produce one hunk
	lines := []Line{
		{Kind: OpEqual, Content: "a"},
		{Kind: OpDelete, Content: "b"},
		{Kind: OpEqual, Content: "c"},
	}
	hunks := ComputeHunks(lines, 100)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if len(hunks[0].Lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(hunks[0].Lines))
	}
}

func TestComputeHunksAllInserted(t *testing.T) {
	lines := []Line{
		{Kind: OpInsert, Content: "a"},
		{Kind: OpInsert, Content: "b"},
		{Kind: OpInsert, Content: "c"},
	}
	hunks := ComputeHunks(lines, 3)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].OldStart != 1 || hunks[0].NewStart != 1 {
		t.Errorf("expected start (1,1), got (%d,%d)", hunks[0].OldStart, hunks[0].NewStart)
	}
}

func TestComputeHunksAllDeleted(t *testing.T) {
	lines := []Line{
		{Kind: OpDelete, Content: "a"},
		{Kind: OpDelete, Content: "b"},
		{Kind: OpDelete, Content: "c"},
	}
	hunks := ComputeHunks(lines, 3)
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].OldStart != 1 || hunks[0].NewStart != 1 {
		t.Errorf("expected start (1,1), got (%d,%d)", hunks[0].OldStart, hunks[0].NewStart)
	}
}
