package align

import (
	"testing"

	"github.com/amterp/go-delta/internal/diff"
)

func TestAnnotateHunksSimilarLinesPair(t *testing.T) {
	h := diff.Hunk{
		OldStart: 1,
		NewStart: 1,
		Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: `"age": 30,`},
			{Kind: diff.OpInsert, Content: `"age": 31,`},
		},
	}
	annotated := AnnotateHunks([]diff.Hunk{h})
	if len(annotated) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(annotated))
	}
	if len(annotated[0].Pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(annotated[0].Pairs))
	}
	pair := annotated[0].Pairs[0]
	if pair.OldIdx != 0 || pair.NewIdx != 1 {
		t.Errorf("expected pair (0,1), got (%d,%d)", pair.OldIdx, pair.NewIdx)
	}
}

func TestAnnotateHunksDissimilarDontPair(t *testing.T) {
	h := diff.Hunk{
		OldStart: 1,
		NewStart: 1,
		Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: "completely different line"},
			{Kind: diff.OpInsert, Content: "!@#$%^&*()"},
		},
	}
	annotated := AnnotateHunks([]diff.Hunk{h})
	if len(annotated[0].Pairs) != 0 {
		t.Errorf("dissimilar lines should not pair, got %d pairs", len(annotated[0].Pairs))
	}
}

func TestAnnotateHunksGreedyOrdering(t *testing.T) {
	// Two removes followed by two adds. The greedy approach should
	// pair the first remove with the first compatible add.
	h := diff.Hunk{
		OldStart: 1,
		NewStart: 1,
		Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: "foo bar"},
			{Kind: diff.OpDelete, Content: "baz qux"},
			{Kind: diff.OpInsert, Content: "foo baz"},
			{Kind: diff.OpInsert, Content: "baz quux"},
		},
	}
	annotated := AnnotateHunks([]diff.Hunk{h})
	pairs := annotated[0].Pairs

	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}

	// First remove ("foo bar") should pair with first add ("foo baz")
	if pairs[0].OldIdx != 0 || pairs[0].NewIdx != 2 {
		t.Errorf("pair 0: expected (0,2), got (%d,%d)", pairs[0].OldIdx, pairs[0].NewIdx)
	}
	// Second remove ("baz qux") should pair with second add ("baz quux")
	if pairs[1].OldIdx != 1 || pairs[1].NewIdx != 3 {
		t.Errorf("pair 1: expected (1,3), got (%d,%d)", pairs[1].OldIdx, pairs[1].NewIdx)
	}
}

func TestAnnotateHunksStopsAtContextBoundary(t *testing.T) {
	// A context line between a remove and add should prevent pairing
	h := diff.Hunk{
		OldStart: 1,
		NewStart: 1,
		Lines: []diff.Line{
			{Kind: diff.OpDelete, Content: "foo bar"},
			{Kind: diff.OpEqual, Content: "context"},
			{Kind: diff.OpInsert, Content: "foo baz"},
		},
	}
	annotated := AnnotateHunks([]diff.Hunk{h})
	if len(annotated[0].Pairs) != 0 {
		t.Errorf("context boundary should prevent pairing, got %d pairs", len(annotated[0].Pairs))
	}
}

func TestAnnotateHunksMultipleHunks(t *testing.T) {
	hunks := []diff.Hunk{
		{
			OldStart: 1, NewStart: 1,
			Lines: []diff.Line{
				{Kind: diff.OpDelete, Content: "hello world"},
				{Kind: diff.OpInsert, Content: "hello earth"},
			},
		},
		{
			OldStart: 10, NewStart: 10,
			Lines: []diff.Line{
				{Kind: diff.OpDelete, Content: "foo bar"},
				{Kind: diff.OpInsert, Content: "foo baz"},
			},
		},
	}
	annotated := AnnotateHunks(hunks)
	if len(annotated) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(annotated))
	}
	if len(annotated[0].Pairs) != 1 {
		t.Errorf("hunk 0: expected 1 pair, got %d", len(annotated[0].Pairs))
	}
	if len(annotated[1].Pairs) != 1 {
		t.Errorf("hunk 1: expected 1 pair, got %d", len(annotated[1].Pairs))
	}
}
