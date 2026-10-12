//go:build hyperscan && cgo

package hyperscan

import (
	"github.com/andreyfesunov/waap_regex_bench/internal/baseline"
	"github.com/andreyfesunov/waap_regex_bench/internal/bench"
	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
	"reflect"
	"sort"
	"testing"
)

func TestByteMatching(t *testing.T) {
	sigs := []signatures.Signature{{ID: 10, Expression: `^.$`}, {ID: 20, Expression: `a`}, {ID: 30, Expression: `^$`}, {ID: 40, Expression: `foo.bar`, Flags: "is"}, {ID: 50, Expression: `x\x00y`}, {ID: 60, Expression: `foo$`}, {ID: 70, Expression: `(?=foo)`}}
	e, accepted, rejected, err := Compile(sigs)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if len(accepted) != 6 || len(rejected) != 1 || rejected[0].ID != 70 {
		t.Fatalf("coverage: %v %v", accepted, rejected)
	}
	w, err := e.NewWorker()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	var dst []int
	for _, tc := range []struct {
		data []byte
		want []int
	}{
		{[]byte("aaaa"), []int{20}}, {[]byte{}, []int{30}}, {[]byte("FOO\nBAR"), []int{40}}, {[]byte("x\x00y"), []int{50}}, {[]byte{0xff}, []int{10}}, {[]byte("é"), nil}, {[]byte("foo\n"), []int{60}}, {[]byte("no match"), []int{20}},
	} {
		dst, err = w.Match(tc.data, dst)
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(dst)
		if !reflect.DeepEqual(append([]int(nil), dst...), tc.want) {
			t.Fatalf("%q got %v want %v", tc.data, dst, tc.want)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Match([]byte("a"), nil); err == nil {
		t.Fatal("closed worker accepted scan")
	}
}
func TestParallelAndGoDifferences(t *testing.T) {
	sigs := []signatures.Signature{{ID: 42, Service: "one", Kind: "match", Expression: `^.$`}, {ID: 55, Service: "end", Kind: "match", Expression: `foo$`}}
	e, accepted, _, err := Compile(sigs)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	records := make([]corpus.Record, 1024)
	for i := range records {
		records[i] = corpus.Record{Index: i, Data: []byte("é")}
	}
	factory := func() (bench.Worker, error) {
		w, err := e.NewWorker()
		if err != nil {
			return nil, err
		}
		return w, nil
	}
	a, _, err := bench.Run(records, accepted, 1, factory)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := bench.Run(records, accepted, 4, factory)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("parallel results differ")
	}
	g, _, _ := baseline.Compile(accepted)
	goIDs, _ := g.NewWorker().Match([]byte("é"), nil)
	if len(goIDs) != 1 || len(a[0].Services) != 0 {
		t.Fatal("byte/Unicode difference not detected")
	}
}
