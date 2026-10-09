package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
)

// ExplainNode is one row in an indented EXPLAIN tree for the page.
type ExplainNode struct {
	Node     string        `json:"node"`
	Relation string        `json:"relation,omitempty"`
	Cost     string        `json:"cost,omitempty"`
	Rows     string        `json:"rows,omitempty"`
	Time     string        `json:"time,omitempty"`
	Depth    int           `json:"depth"`
	Children []ExplainNode `json:"children,omitempty"`
}

// ExplainRequest runs EXPLAIN for one statement.
type ExplainRequest struct {
	SQL     string
	Analyze bool
	Confirm bool
}

// ExplainResult is the parsed plan tree.
type ExplainResult struct {
	Root ExplainNode `json:"root"`
	Raw  string      `json:"-"`
}

// Explain runs EXPLAIN (FORMAT JSON) and parses the plan.
func (s *Session) Explain(ctx context.Context, req ExplainRequest) (ExplainResult, error) {
	sqlText := strings.TrimSpace(req.SQL)
	if sqlText == "" {
		return ExplainResult{}, errors.New("session: sql required")
	}
	explainSQL, err := buildExplainSQL(s.profile.Engine, sqlText, req.Analyze)
	if err != nil {
		return ExplainResult{}, err
	}
	check, err := s.Guard(explainSQL, req.Confirm)
	if err != nil {
		return ExplainResult{}, err
	}
	if check == WriteNeedsConfirmation {
		return ExplainResult{}, errExplainNeedsConfirm
	}
	if check == WriteBlocked {
		return ExplainResult{}, ErrWriteBlockedReadOnly()
	}
	conn := s.mainConn()
	if conn == nil {
		return ExplainResult{}, errSessionNotReady
	}
	rows, err := conn.Query(ctx, explainSQL)
	if err != nil {
		return ExplainResult{}, err
	}
	defer func() { _ = rows.Close() }()
	w, err := rows.NextWindow(1)
	if err != nil {
		return ExplainResult{}, err
	}
	if w == nil || w.Rows == 0 {
		return ExplainResult{}, errors.New("session: empty explain result")
	}
	raw := cellText(w, 0, 0)
	root, err := parseExplainJSON(s.profile.Engine, raw)
	if err != nil {
		return ExplainResult{}, err
	}
	return ExplainResult{Root: root, Raw: raw}, nil
}

var errExplainNeedsConfirm = errors.New("session: explain analyze needs confirmation")

// ErrExplainNeedsConfirm reports safe-mode / analyze confirmation.
func ErrExplainNeedsConfirm() error { return errExplainNeedsConfirm }

func buildExplainSQL(engine, inner string, analyze bool) (string, error) {
	eng := strings.ToLower(engine)
	switch eng {
	case "postgres", "":
		if analyze {
			return fmt.Sprintf("EXPLAIN (ANALYZE, FORMAT JSON) %s", inner), nil
		}
		return fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", inner), nil
	case "mysql", "mariadb":
		if analyze {
			return fmt.Sprintf("EXPLAIN ANALYZE FORMAT=JSON %s", inner), nil
		}
		return fmt.Sprintf("EXPLAIN FORMAT=JSON %s", inner), nil
	default:
		return "", fmt.Errorf("session: explain not supported for %q", engine)
	}
}

func cellText(w *driver.Window, row, col int) string {
	if w == nil || row >= w.Rows || col >= len(w.Columns) {
		return ""
	}
	s, ok := w.CellText(col, row)
	if !ok {
		return ""
	}
	return s
}

func parseExplainJSON(engine, raw string) (ExplainNode, error) {
	eng := strings.ToLower(engine)
	if eng == "mysql" || eng == "mariadb" {
		return parseMySQLExplainJSON(raw)
	}
	return parsePostgresExplainJSON(raw)
}

func parsePostgresExplainJSON(raw string) (ExplainNode, error) {
	var top []struct {
		Plan json.RawMessage `json:"Plan"`
	}
	if err := json.Unmarshal([]byte(raw), &top); err != nil {
		return ExplainNode{}, err
	}
	if len(top) == 0 {
		return ExplainNode{}, errors.New("session: no plan")
	}
	return decodePGPlan(top[0].Plan, 0)
}

type pgPlan struct {
	NodeType        string   `json:"Node Type"`
	RelationName    string   `json:"Relation Name"`
	Alias           string   `json:"Alias"`
	StartupCost     float64  `json:"Startup Cost"`
	TotalCost       float64  `json:"Total Cost"`
	PlanRows        float64  `json:"Plan Rows"`
	ActualTotalTime float64  `json:"Actual Total Time"`
	Plans           []pgPlan `json:"Plans"`
}

func decodePGPlan(raw json.RawMessage, depth int) (ExplainNode, error) {
	var p pgPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		return ExplainNode{}, err
	}
	n := ExplainNode{
		Node:  p.NodeType,
		Depth: depth,
		Cost:  fmt.Sprintf("%.2f..%.2f", p.StartupCost, p.TotalCost),
		Rows:  fmt.Sprintf("%.0f", p.PlanRows),
	}
	if p.RelationName != "" {
		n.Relation = p.RelationName
		if p.Alias != "" && p.Alias != p.RelationName {
			n.Relation = p.RelationName + " " + p.Alias
		}
	}
	if p.ActualTotalTime > 0 {
		n.Time = fmt.Sprintf("%.3f ms", p.ActualTotalTime)
	}
	for _, child := range p.Plans {
		craw, _ := json.Marshal(child)
		ch, err := decodePGPlan(craw, depth+1)
		if err != nil {
			return ExplainNode{}, err
		}
		n.Children = append(n.Children, ch)
	}
	return n, nil
}

func parseMySQLExplainJSON(raw string) (ExplainNode, error) {
	var doc struct {
		QueryBlock struct {
			Plan json.RawMessage `json:"plan"`
		} `json:"query_block"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return ExplainNode{}, err
	}
	if len(doc.QueryBlock.Plan) == 0 {
		return ExplainNode{}, errors.New("session: no plan")
	}
	return decodeMySQLPlan(doc.QueryBlock.Plan, 0)
}

func decodeMySQLPlan(raw json.RawMessage, depth int) (ExplainNode, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return ExplainNode{}, err
	}
	n := ExplainNode{Depth: depth}
	if op, ok := m["operation"]; ok {
		_ = json.Unmarshal(op, &n.Node)
	}
	if table, ok := m["table_name"]; ok {
		_ = json.Unmarshal(table, &n.Relation)
	}
	if cost, ok := m["cost"]; ok {
		var c struct {
			Read float64 `json:"read"`
		}
		_ = json.Unmarshal(cost, &c)
		n.Cost = fmt.Sprintf("%.2f", c.Read)
	}
	if rows, ok := m["rows_examined_per_scan"]; ok {
		_ = json.Unmarshal(rows, &n.Rows)
	}
	if nested, ok := m["nested_loop"]; ok {
		var parts []json.RawMessage
		_ = json.Unmarshal(nested, &parts)
		for _, part := range parts {
			ch, err := decodeMySQLPlan(part, depth+1)
			if err != nil {
				return ExplainNode{}, err
			}
			n.Children = append(n.Children, ch)
		}
	}
	return n, nil
}
