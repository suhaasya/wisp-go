package bench

// Budget is one row from the PRD gate tables (target = goal, cap = deal breaker).
type Budget struct {
	Metric string
	Target float64
	Cap    float64
	Unit   string
}

// GoBudgets is the Go process gate table (single source of truth).
var GoBudgets = []Budget{
	{Metric: "binary_mb", Target: 15, Cap: 25, Unit: "MB"},
	{Metric: "download_mb", Target: 7, Cap: 10, Unit: "MB"},
	{Metric: "idle_rss_mb", Target: 20, Cap: 30, Unit: "MB"},
	{Metric: "one_conn_500_rss_mb", Target: 30, Cap: 45, Unit: "MB"},
	{Metric: "per_conn_rss_mb", Target: 2, Cap: 4, Unit: "MB"},
	{Metric: "scroll_1m_rss_mb", Target: 50, Cap: 80, Unit: "MB"},
	{Metric: "start_ms", Target: 100, Cap: 300, Unit: "ms"},
}

// TabBudgets is the browser tab gate table (single source of truth).
var TabBudgets = []Budget{
	{Metric: "js_heap_idle_mb", Target: 5, Cap: 10, Unit: "MB"},
	{Metric: "js_heap_1m_mb", Target: 10, Cap: 20, Unit: "MB"},
	{Metric: "dom_nodes", Target: 2000, Cap: 4000, Unit: "nodes"},
	{Metric: "js_initial_kb_gz", Target: 60, Cap: 100, Unit: "KB"},
	{Metric: "js_editor_kb_gz", Target: 200, Cap: 280, Unit: "KB"},
	{Metric: "css_kb_gz", Target: 15, Cap: 25, Unit: "KB"},
	{Metric: "tab_total_mb", Target: 40, Cap: 70, Unit: "MB"},
	{Metric: "long_task_ms", Target: 50, Cap: 100, Unit: "ms"},
}
