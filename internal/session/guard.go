package session

import (
	"errors"
	"strings"
	"unicode"
)

var errWriteBlockedReadOnly = errors.New("session: write blocked (read-only)")

// ErrWriteBlockedReadOnly is returned when a write is rejected on a read-only session.
func ErrWriteBlockedReadOnly() error { return errWriteBlockedReadOnly }

// ClassifyStatement returns read vs write using the first keyword after comments.
func ClassifyStatement(sql string) StatementClass {
	s := stripSQLComments(strings.TrimSpace(sql))
	if s == "" {
		return StatementWrite
	}
	kw, rest := firstKeyword(s)
	switch kw {
	case "SELECT", "SHOW", "DESCRIBE", "VALUES":
		return StatementRead
	case "EXPLAIN":
		if explainHasAnalyze(rest) {
			return StatementWrite
		}
		return StatementRead
	case "WITH":
		if withStatementIsRead(rest) {
			return StatementRead
		}
		return StatementWrite
	default:
		return StatementWrite
	}
}

// Guard enforces read-only and safe-mode rules before SQL is sent to the database.
func (s *Session) Guard(sql string, confirmed bool) (WriteCheck, error) {
	if s == nil {
		return WriteBlocked, errors.New("session: nil")
	}
	st := s.State()
	if st != StateReady && st != StateBusy {
		return WriteBlocked, errors.New("session: not ready")
	}
	if ClassifyStatement(sql) == StatementRead {
		return WriteAllowed, nil
	}
	if s.profile.ReadOnly {
		return WriteBlocked, errWriteBlockedReadOnly
	}
	if s.profile.SafeMode && !confirmed {
		return WriteNeedsConfirmation, nil
	}
	if IsDDL(sql) {
		s.InvalidateMetadata()
	}
	return WriteAllowed, nil
}

// IsDDL reports whether sql is a data-definition statement.
func IsDDL(sql string) bool {
	s := stripSQLComments(strings.TrimSpace(sql))
	if s == "" {
		return false
	}
	kw, _ := firstKeyword(s)
	switch kw {
	case "CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME":
		return true
	default:
		return false
	}
}

func stripSQLComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '-' && s[i+1] == '-' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			i += 2
			for i+1 < len(s) && (s[i] != '*' || s[i+1] != '/') {
				i++
			}
			if i+1 < len(s) {
				i += 2
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func firstKeyword(s string) (keyword, rest string) {
	s = strings.TrimSpace(s)
	i := 0
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}
	start := i
	for i < len(s) && !unicode.IsSpace(rune(s[i])) {
		i++
	}
	if start >= len(s) {
		return "", ""
	}
	keyword = strings.ToUpper(s[start:i])
	rest = strings.TrimSpace(s[i:])
	return keyword, rest
}

func explainHasAnalyze(rest string) bool {
	up := strings.ToUpper(rest)
	return strings.Contains(up, " ANALYZE") || strings.HasPrefix(up, "ANALYZE")
}

func withStatementIsRead(rest string) bool {
	up := strings.ToUpper(rest)
	for _, w := range []string{"INSERT", "UPDATE", "DELETE", "MERGE", "REPLACE"} {
		if strings.Contains(up, w) {
			return false
		}
	}
	return strings.Contains(up, "SELECT")
}

const (
	pgReadOnlySQL  = "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY"
	myReadOnlySQL  = "SET SESSION TRANSACTION READ ONLY"
	defaultKeepSec = 30
)

// ReadOnlySessionSQL returns the engine-specific read-only session statement.
func ReadOnlySessionSQL(engine string) string {
	return readOnlySessionSQL(engine)
}

func readOnlySessionSQL(engine string) string {
	switch engine {
	case "postgres":
		return pgReadOnlySQL
	case "mysql", "mariadb":
		return myReadOnlySQL
	default:
		return ""
	}
}
