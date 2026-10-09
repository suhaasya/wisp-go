package bench

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const modulePath = "github.com/suhaasya/wisp-go"

// ImportViolation records a disallowed edge between internal packages.
type ImportViolation struct {
	Importer string
	Imported string
}

// allowedInternalDeps maps each internal package import path to the internal
// packages it may import directly (AGENTS.md section 2).
var allowedInternalDeps = map[string][]string{
	modulePath + "/internal/transport":       {},
	modulePath + "/internal/transport/ssh":   {},
	modulePath + "/internal/driver":          {modulePath + "/internal/transport"},
	modulePath + "/internal/driver/postgres": {modulePath + "/internal/driver"},
	modulePath + "/internal/driver/mysql":    {modulePath + "/internal/driver"},
	modulePath + "/internal/driver/mock":     {modulePath + "/internal/driver"},
	modulePath + "/internal/pager":           {modulePath + "/internal/driver", modulePath + "/internal/sqlgen"},
	modulePath + "/internal/sqlgen":          {modulePath + "/internal/driver"},
	modulePath + "/internal/store":           {},
	modulePath + "/internal/session": {
		modulePath + "/internal/mover",
		modulePath + "/internal/pager",
		modulePath + "/internal/sqlgen",
		modulePath + "/internal/store",
		modulePath + "/internal/transport",
		modulePath + "/internal/transport/ssh",
		modulePath + "/internal/driver",
		modulePath + "/internal/driver/postgres",
		modulePath + "/internal/driver/mysql",
	},
	modulePath + "/internal/api":    {modulePath + "/internal/session", modulePath + "/internal/store"},
	modulePath + "/internal/server": {modulePath + "/internal/api"},
	modulePath + "/internal/mover":  {modulePath + "/internal/driver", modulePath + "/internal/sqlgen"},
	modulePath + "/internal/bench":  {},
	modulePath + "/internal/update": {},
}

// CheckInternalImportRules runs go list over ./internal/... and returns violations.
func CheckInternalImportRules() ([]ImportViolation, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "go", "list", "-f", "{{.ImportPath}}\t{{join .Imports \"\t\"}}", "./internal/...")
	cmd.Dir = moduleRoot()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	var violations []ImportViolation
	for line := range bytes.SplitSeq(out, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(string(line), "\t")
		pkg := parts[0]
		allowed := allowedSet(allowedInternalDeps[pkg])
		for _, imp := range parts[1:] {
			if imp == "" || !strings.HasPrefix(imp, modulePath+"/internal/") {
				continue
			}
			if !allowed[imp] {
				violations = append(violations, ImportViolation{Importer: pkg, Imported: imp})
			}
		}
	}
	return violations, nil
}

func allowedSet(list []string) map[string]bool {
	m := make(map[string]bool, len(list))
	for _, p := range list {
		m[p] = true
	}
	return m
}

func moduleRoot() string {
	ctx := context.Background()
	out, err := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err != nil {
		return "."
	}
	return strings.TrimSpace(string(out))
}
