package indexdb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexProjectCodeFiles_WritesPostings(t *testing.T) {
	projectDir := t.TempDir()
	dbDir := filepath.Join(t.TempDir(), "indexdb")
	_ = os.Setenv("IAT_INDEXDB_DIR", dbDir)
	defer func() { _ = CloseDefault() }()

	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("# hello\nindex me\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := IndexProjectCodeFiles(1, projectDir, []string{"main.go", "README.md"})
	if err != nil {
		t.Fatalf("index failed: %v", err)
	}
	t.Logf("Indexed info: Files=%d, TotalTokens=%d", info.Files, info.TotalTokens)
	if info.Files == 0 {
		t.Fatalf("expected files > 0, got %d", info.Files)
	}

	db, err := OpenDefault()
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}

	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	foundPosting := false
	for iter.Next() {
		k := string(iter.Key())
		if len(k) >= len("tokf:1:") && k[:len("tokf:1:")] == "tokf:1:" {
			foundPosting = true
			break
		}
	}
	if err := iter.Error(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if !foundPosting {
		t.Fatalf("expected tokf:1:* keys to be written")
	}
}
