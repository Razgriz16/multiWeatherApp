package database

import (
	"path/filepath"
	"testing"
)

func TestConditionIDs(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ids, err := ConditionIDs(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 29 {
		t.Fatalf("expected 29 seeded conditions, got %d", len(ids))
	}
	if _, ok := ids[95]; !ok {
		t.Errorf("expected code 95 (thunderstorm) to be seeded")
	}
}
