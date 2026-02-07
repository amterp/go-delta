package align

// AlignOp classifies a token in an alignment.
type AlignOp int

const (
	AlignMatch  AlignOp = iota // token is unchanged
	AlignDelete                // token exists only in old
	AlignInsert                // token exists only in new
)

// AlignedToken pairs an alignment operation with the token it applies to.
type AlignedToken struct {
	Op    AlignOp
	Token Token
}

// Alignment holds the result of aligning two token sequences.
type Alignment struct {
	Old      []AlignedToken // aligned old tokens
	New      []AlignedToken // aligned new tokens
	Distance float64        // normalized edit distance [0, 1]
}

// Align performs Needleman-Wunsch alignment on two token sequences.
// Scoring: match=0, mismatch=1, gap=1. Returns the alignment with
// a normalized distance suitable for deciding whether two lines are
// similar enough to pair.
func Align(oldTokens, newTokens []Token) Alignment {
	n := len(oldTokens)
	m := len(newTokens)

	if n == 0 && m == 0 {
		return Alignment{}
	}
	if n == 0 {
		aligned := make([]AlignedToken, m)
		for i, t := range newTokens {
			aligned[i] = AlignedToken{Op: AlignInsert, Token: t}
		}
		return Alignment{
			New:      aligned,
			Distance: 1.0,
		}
	}
	if m == 0 {
		aligned := make([]AlignedToken, n)
		for i, t := range oldTokens {
			aligned[i] = AlignedToken{Op: AlignDelete, Token: t}
		}
		return Alignment{
			Old:      aligned,
			Distance: 1.0,
		}
	}

	// DP matrix: dp[i][j] = cost of aligning old[:i] with new[:j]
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		dp[i][0] = i
	}
	for j := 1; j <= m; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1 // mismatch
			if oldTokens[i-1].Text == newTokens[j-1].Text {
				cost = 0 // match
			}
			del := dp[i-1][j] + 1
			ins := dp[i][j-1] + 1
			sub := dp[i-1][j-1] + cost

			dp[i][j] = min(del, min(ins, sub))
		}
	}

	// Traceback
	var oldAligned, newAligned []AlignedToken
	i, j := n, m
	matches, deletions, insertions := 0, 0, 0

	for i > 0 || j > 0 {
		if i > 0 && dp[i][j] == dp[i-1][j]+1 {
			// deletion - checked before match so that deletions
			// in runs of identical tokens land at the end of the
			// run rather than clustering at the start
			oldAligned = append(oldAligned, AlignedToken{Op: AlignDelete, Token: oldTokens[i-1]})
			deletions++
			i--
		} else if j > 0 && dp[i][j] == dp[i][j-1]+1 {
			// insertion - same reasoning as deletion
			newAligned = append(newAligned, AlignedToken{Op: AlignInsert, Token: newTokens[j-1]})
			insertions++
			j--
		} else if i > 0 && j > 0 && oldTokens[i-1].Text == newTokens[j-1].Text &&
			dp[i][j] == dp[i-1][j-1] {
			// match
			oldAligned = append(oldAligned, AlignedToken{Op: AlignMatch, Token: oldTokens[i-1]})
			newAligned = append(newAligned, AlignedToken{Op: AlignMatch, Token: newTokens[j-1]})
			matches++
			i--
			j--
		} else if i > 0 && j > 0 &&
			dp[i][j] == dp[i-1][j-1]+1 {
			// substitution (mismatch): treat as delete + insert
			oldAligned = append(oldAligned, AlignedToken{Op: AlignDelete, Token: oldTokens[i-1]})
			newAligned = append(newAligned, AlignedToken{Op: AlignInsert, Token: newTokens[j-1]})
			deletions++
			insertions++
			i--
			j--
		} else {
			// fallback: j > 0 is guaranteed here because the DP
			// boundary values ensure the delete/insert branches
			// above catch all i>0,j==0 and i==0,j>0 cases.
			newAligned = append(newAligned, AlignedToken{Op: AlignInsert, Token: newTokens[j-1]})
			insertions++
			j--
		}
	}

	// Reverse to forward order
	reverse(oldAligned)
	reverse(newAligned)

	// Normalized distance: (del + ins) / (2*matches + del + ins)
	total := 2*matches + deletions + insertions
	var dist float64
	if total > 0 {
		dist = float64(deletions+insertions) / float64(total)
	}

	return Alignment{
		Old:      oldAligned,
		New:      newAligned,
		Distance: dist,
	}
}

func reverse(s []AlignedToken) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
