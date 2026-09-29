package database

import (
	"path/filepath"
	"testing"
)

func TestAPIModels(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	models, err := APIModels(db)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ecmwf_ifs025", "gfs_seamless", "icon_seamless"}
	if len(models) != len(want) {
		t.Fatalf("got %v, want %v", models, want)
	}
	for i := range want {
		if models[i] != want[i] {
			t.Errorf("got %v, want %v", models, want)
			break
		}
	}
}
