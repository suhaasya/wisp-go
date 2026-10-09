package session_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestClassifyStatementReadKeywords(t *testing.T) {
	reads := []string{
		"SELECT 1",
		"  select * from t",
		"-- comment\nSELECT 1",
		"SHOW TABLES",
		"DESCRIBE users",
		"VALUES (1,2)",
		"EXPLAIN SELECT 1",
		"EXPLAIN (SELECT 1)",
		"WITH c AS (SELECT 1) SELECT * FROM c",
	}
	for _, sql := range reads {
		if session.ClassifyStatement(sql) != session.StatementRead {
			t.Fatalf("%q want read", sql)
		}
	}
}

func TestClassifyStatementWriteKeywords(t *testing.T) {
	writes := []string{
		"INSERT INTO t VALUES (1)",
		"UPDATE t SET x=1",
		"DELETE FROM t",
		"CREATE TABLE t(id int)",
		"EXPLAIN ANALYZE SELECT 1",
		"WITH c AS (INSERT INTO t SELECT 1 RETURNING id) SELECT * FROM c",
		"WITH c AS (SELECT 1) INSERT INTO t SELECT * FROM c",
	}
	for _, sql := range writes {
		if session.ClassifyStatement(sql) != session.StatementWrite {
			t.Fatalf("%q want write", sql)
		}
	}
}

func TestGuardReadOnlyBlocksWrite(t *testing.T) {
	s := session.NewTestSession(store.Profile{ReadOnly: true, Engine: "postgres"})
	check, err := s.Guard("INSERT INTO t VALUES (1)", false)
	if err == nil || check != session.WriteBlocked {
		t.Fatalf("check=%v err=%v", check, err)
	}
	if !session.IsWriteBlockedReadOnly(err) {
		t.Fatalf("want read-only error, got %v", err)
	}
}

func TestGuardReadOnlyAllowsSelect(t *testing.T) {
	s := session.NewTestSession(store.Profile{ReadOnly: true})
	check, err := s.Guard("SELECT 1", false)
	if err != nil || check != session.WriteAllowed {
		t.Fatalf("check=%v err=%v", check, err)
	}
}

func TestGuardSafeModeNeedsConfirmation(t *testing.T) {
	s := session.NewTestSession(store.Profile{SafeMode: true})
	check, err := s.Guard("UPDATE t SET x=1", false)
	if err != nil || check != session.WriteNeedsConfirmation {
		t.Fatalf("check=%v err=%v", check, err)
	}
	check, err = s.Guard("UPDATE t SET x=1", true)
	if err != nil || check != session.WriteAllowed {
		t.Fatalf("confirmed: check=%v err=%v", check, err)
	}
}
