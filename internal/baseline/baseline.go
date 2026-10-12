package baseline

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
)

type Rejection struct {
	ID     int    `json:"id"`
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

type rule struct {
	id    int
	regex *regexp.Regexp
}
type Engine struct{ rules []rule }

func Compile(input []signatures.Signature) (*Engine, []signatures.Signature, []Rejection) {
	e := &Engine{}
	var accepted []signatures.Signature
	var rejected []Rejection
	for _, s := range input {
		var err error
		for _, f := range s.Flags {
			if f != 'i' && f != 's' {
				err = fmt.Errorf("unsupported flag %q", f)
				break
			}
		}
		expr := s.Expression
		if s.Flags != "" {
			expr = "(?" + s.Flags + ")" + expr
		}
		var re *regexp.Regexp
		if err == nil {
			re, err = regexp.Compile(expr)
		}
		if err != nil {
			rejected = append(rejected, Rejection{s.ID, s.Line, strings.TrimSpace(err.Error())})
			continue
		}
		e.rules = append(e.rules, rule{s.ID, re})
		accepted = append(accepted, s)
	}
	return e, accepted, rejected
}

// Worker owns its caller-provided result buffer; compiled rules are immutable.
type Worker struct{ engine *Engine }

func (e *Engine) NewWorker() *Worker { return &Worker{engine: e} }
func (w *Worker) Match(data []byte, dst []int) ([]int, error) {
	dst = dst[:0]
	for _, rule := range w.engine.rules {
		if rule.regex.Match(data) {
			dst = append(dst, rule.id)
		}
	}
	return dst, nil
}
