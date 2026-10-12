package bench

import (
	"errors"
	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	"testing"
)

type resourceWorker struct {
	closed    int
	fail      bool
	closeFail bool
}

func (w *resourceWorker) Match(_ []byte, dst []int) ([]int, error) {
	if w.fail {
		return nil, errors.New("scan failed")
	}
	return dst[:0], nil
}
func (w *resourceWorker) Close() error {
	w.closed++
	if w.closeFail {
		return errors.New("close failed")
	}
	return nil
}
func TestResourceCleanup(t *testing.T) {
	for _, mode := range []string{"ok", "scan", "factory", "close"} {
		var made []*resourceWorker
		_, _, err := Run([]corpus.Record{{Data: []byte("x")}}, nil, 3, func() (Worker, error) {
			if mode == "factory" && len(made) == 2 {
				return nil, errors.New("factory failed")
			}
			w := &resourceWorker{fail: mode == "scan", closeFail: mode == "close"}
			made = append(made, w)
			return w, nil
		})
		if (mode == "ok") != (err == nil) {
			t.Fatalf("mode %s: %v", mode, err)
		}
		for _, w := range made {
			if w.closed != 1 {
				t.Fatalf("mode %s: closed %d times", mode, w.closed)
			}
		}
	}
}
