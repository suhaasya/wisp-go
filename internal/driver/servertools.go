package driver

// ProcessHighlight values for the wisp_highlight column in server tools process lists.
const (
	HighlightBlocked     = "blocked"
	HighlightLongRunning = "long_running"
)

// ClassifyProcessHighlight returns the wisp_highlight cell for a process row.
func ClassifyProcessHighlight(seconds int, blocked bool) string {
	if blocked {
		return HighlightBlocked
	}
	if seconds > 30 {
		return HighlightLongRunning
	}
	return ""
}
