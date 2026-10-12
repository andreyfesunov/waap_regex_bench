package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	r := Request{Method: "GET", URL: "/?x=%2f", Headers: map[string]string{"z": "Z", "A": "a"}, Data: "body"}
	want := "GET /?x=%2f\r\nA: a\r\nz: Z\r\n\r\nbody"
	if got := string(Normalize(r)); got != want {
		t.Fatalf("got %q", got)
	}
	r.Headers = map[string]string{"A": "a", "z": "Z"}
	if string(Normalize(r)) != want {
		t.Fatal("unstable header order")
	}
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"legitimate", "malicious"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
func write(t *testing.T, root, path, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, path), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

const row = `{"method":"GET","url":"/","headers":{},"data":""}`

func TestLoadOrderAndLimit(t *testing.T) {
	root := fixture(t)
	write(t, root, "legitimate/z.json", "["+row+"]")
	write(t, root, "legitimate/a.json", "["+row+","+row+"]")
	write(t, root, "malicious/m.json", "["+row+"]")
	all, err := Load(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"legitimate/a.json", "malicious/m.json", "legitimate/a.json", "legitimate/z.json"}
	if len(all) != len(want) {
		t.Fatalf("got %d", len(all))
	}
	for i, r := range all {
		if r.Source != want[i] {
			t.Fatalf("record %d: %s", i, r.Source)
		}
	}
	if all[2].Index != 1 {
		t.Fatal("wrong source index")
	}
	limited, err := Load(root, 2)
	if err != nil || len(limited) != 2 {
		t.Fatalf("limit %v %v", limited, err)
	}
}

func TestLoadErrors(t *testing.T) {
	for _, bad := range []string{`[{"method":1,"url":"/","headers":{},"data":""}]`, `[{"method":"GET","url":"/","headers":{"X":null},"data":""}]`, "[" + row + ",", "[" + row + "] {}"} {
		root := fixture(t)
		write(t, root, "legitimate/a.json", bad)
		write(t, root, "malicious/a.json", "[]")
		_, err := Load(root, 0)
		if err == nil || !strings.Contains(err.Error(), "legitimate/a.json record") {
			t.Fatalf("missing source context: %v", err)
		}
	}
}
