package bench

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andreyfesunov/waap_regex_bench/internal/baseline"
	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
)

func TestWorkerCountsAndPriority(t *testing.T) {
	sigs := []signatures.Signature{
		{ID: 0, Service: "http", Kind: "softmatch", Probe: "first", ProbeOrder: 0, Line: 1, Expression: "x"},
		{ID: 1, Service: "http", Kind: "match", Probe: "later", ProbeOrder: 1, Line: 2, Expression: "x"},
		{ID: 2, Service: "http", Kind: "match", Probe: "first", ProbeOrder: 0, Line: 4, Expression: "x"},
		{ID: 3, Service: "http", Kind: "match", Probe: "first", ProbeOrder: 0, Line: 3, Expression: "x"},
		{ID: 4, Service: "aaa", Kind: "softmatch", Probe: "first", ProbeOrder: 0, Line: 5, Expression: "x"},
	}
	e, accepted, _ := baseline.Compile(sigs)
	factory := func() (Worker, error) { return e.NewWorker(), nil }
	records := make([]corpus.Record, 1025)
	for i := range records {
		records[i] = corpus.Record{Source: "fixture", Index: i, Data: []byte("x")}
	}
	records[0].Data = []byte("no match")
	one, stats, err := Run(records, accepted, 1, factory)
	if err != nil {
		t.Fatal(err)
	}
	many, multi, err := Run(records, accepted, 4, factory)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(one, many) {
		t.Fatal("parallel results differ")
	}
	if stats.PatternMatches != 5120 || stats.ServiceMatches != 2048 || multi.GOMAXPROCS != 4 {
		t.Fatalf("wrong stats: %+v %+v", stats, multi)
	}
	if len(one[0].Services) != 0 || one[0].Services == nil {
		t.Fatal("empty matches must be []")
	}
	if one[1].Services[0].Name != "aaa" || one[1].Services[1].PatternID != 3 {
		t.Fatalf("wrong priority: %+v", one[1])
	}
	p := filepath.Join(t.TempDir(), "results.jsonl")
	if err := WriteJSONL(p, one[:2]); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil || !strings.Contains(string(b), `"services":[]`) || strings.Count(string(b), "\n") != 2 {
		t.Fatalf("bad JSONL: %s %v", b, err)
	}
}

type failing struct{}

func (failing) Match([]byte, []int) ([]int, error) { return nil, errors.New("scan failure") }
func TestEngineFailure(t *testing.T) {
	_, _, err := Run([]corpus.Record{{Data: []byte("x")}}, nil, 2, func() (Worker, error) { return failing{}, nil })
	if err == nil || !strings.Contains(err.Error(), "scan failure") {
		t.Fatalf("got %v", err)
	}
}

func TestFilteredNoncontiguousIDs(t *testing.T) {
	sigs := []signatures.Signature{{ID: 100, Service: "http", Kind: "match", Expression: "x"}}
	e, accepted, _ := baseline.Compile(sigs)
	r, _, err := Run([]corpus.Record{{Data: []byte("x")}}, accepted, 1, func() (Worker, error) { return e.NewWorker(), nil })
	if err != nil || r[0].Services[0].PatternID != 100 {
		t.Fatalf("filtered IDs: %v %v", r, err)
	}
}
