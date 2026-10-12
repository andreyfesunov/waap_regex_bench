package baseline

import (
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
	"testing"
)

func TestCompileAndExhaustiveMatch(t *testing.T) {
	input := []signatures.Signature{
		{ID: 10, Line: 3, Expression: `foo.bar`, Flags: "is"},
		{ID: 20, Line: 4, Expression: `FOO`},
		{ID: 30, Line: 5, Expression: `(?=foo)`},
		{ID: 40, Line: 6, Expression: `x`, Flags: "q"},
	}
	e, accepted, rejected := Compile(input)
	if len(accepted) != 2 || len(rejected) != 2 || rejected[0].ID != 30 || rejected[0].Line != 5 || rejected[0].Reason == "" {
		t.Fatalf("bad compilation report: %v %v", accepted, rejected)
	}
	w := e.NewWorker()
	ids, err := w.Match([]byte("FOO\nbar"), []int{999})
	if err != nil || len(ids) != 2 || ids[0] != 10 || ids[1] != 20 {
		t.Fatalf("match %v %v", ids, err)
	}
	ids, err = w.Match([]byte("nothing"), ids)
	if err != nil || len(ids) != 0 {
		t.Fatalf("stale matches: %v %v", ids, err)
	}
}
