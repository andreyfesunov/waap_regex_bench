package signatures

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifest(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "patterns.json")
	sigs := []Signature{{ID: 10}, {ID: 20}}
	if err := SaveManifest(path, "hash", "common", sigs); err != nil {
		t.Fatal(err)
	}
	selected, err := SelectManifest(path, "hash", sigs)
	if err != nil || len(selected) != 2 || selected[1].ID != 20 {
		t.Fatalf("%v %v", selected, err)
	}
	if _, err := SelectManifest(path, "changed", sigs); err == nil {
		t.Fatal("accepted wrong hash")
	}
	for _, ids := range []string{"[10,10]", "[99]", "[]"} {
		if err := os.WriteFile(path, []byte(`{"source_sha256":"hash","ids":`+ids+`}`), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := SelectManifest(path, "hash", sigs); err == nil {
			t.Fatalf("accepted %s", ids)
		}
	}
}
