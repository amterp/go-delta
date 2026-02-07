// Package godelta renders human-readable, colored terminal diffs.
//
// Usage:
//
//	import gd "github.com/amterp/go-delta"
//	output := gd.Diff(old, new)
package godelta

// Diff computes and renders a colored diff between two strings.
// Returns an empty string if inputs are identical.
func Diff(old, new string) string {
	return DiffWith(old, new)
}

// DiffWith computes a diff using the provided options.
func DiffWith(old, new string, opts ...Option) string {
	if old == new {
		return ""
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	useColor := resolveColor(cfg.colorMode)
	styles := buildStyles(useColor)

	return runPipeline(old, new, cfg, styles)
}
