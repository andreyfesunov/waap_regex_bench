//go:build hyperscan && cgo

// Package hyperscan implements byte-mode multi-pattern matching.
package hyperscan

import (
	"errors"
	"fmt"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
	hs "github.com/flier/gohs/hyperscan"
)

type Rejection struct {
	ID     int    `json:"id"`
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}
type Engine struct{ db hs.BlockDatabase }

func Compile(input []signatures.Signature) (*Engine, []signatures.Signature, []Rejection, error) {
	if err := hs.ValidPlatform(); err != nil {
		return nil, nil, nil, err
	}
	var patterns []*hs.Pattern
	var accepted []signatures.Signature
	var rejected []Rejection
	for _, s := range input {
		flags := hs.SingleMatch | hs.AllowEmpty
		var err error
		for _, f := range s.Flags {
			switch f {
			case 'i':
				flags |= hs.Caseless
			case 's':
				flags |= hs.DotAll
			default:
				err = fmt.Errorf("unsupported flag %q", f)
			}
		}
		p := hs.NewPattern(s.Expression, flags)
		p.Id = s.ID
		if err == nil {
			_, err = p.Info()
		}
		if err != nil {
			rejected = append(rejected, Rejection{s.ID, s.Line, err.Error()})
			continue
		}
		patterns = append(patterns, p)
		accepted = append(accepted, s)
	}
	for len(patterns) > 0 {
		db, err := hs.NewBlockDatabase(patterns...)
		if err == nil {
			return &Engine{db}, accepted, rejected, nil
		}
		var ce *hs.CompileError
		if !errors.As(err, &ce) || ce.Expression < 0 || ce.Expression >= len(accepted) {
			return nil, nil, rejected, fmt.Errorf("compile database: %w", err)
		}
		i := ce.Expression
		s := accepted[i]
		rejected = append(rejected, Rejection{s.ID, s.Line, err.Error()})
		patterns = append(patterns[:i], patterns[i+1:]...)
		accepted = append(accepted[:i], accepted[i+1:]...)
	}
	return nil, nil, rejected, fmt.Errorf("no compatible Hyperscan patterns")
}
func (e *Engine) Close() error {
	if e.db == nil {
		return nil
	}
	err := e.db.Close()
	e.db = nil
	return err
}
func (e *Engine) Size() (int, error) { return e.db.Size() }
func Version() string                { return hs.Version() }

type Worker struct {
	engine  *Engine
	scratch *hs.Scratch
	ids     []int
	handler hs.MatchHandler
}

func (e *Engine) NewWorker() (*Worker, error) {
	s, err := hs.NewScratch(e.db)
	if err != nil {
		return nil, err
	}
	w := &Worker{engine: e, scratch: s}
	w.handler = func(id uint, from, to uint64, flags uint, context interface{}) error {
		w.ids = append(w.ids, int(id))
		return nil
	}
	return w, nil
}
func (w *Worker) Size() (int, error) { return w.scratch.Size() }
func (w *Worker) Close() error {
	if w.scratch == nil {
		return nil
	}
	err := w.scratch.Free()
	w.scratch = nil
	return err
}
func (w *Worker) Match(data []byte, dst []int) ([]int, error) {
	if w.scratch == nil {
		return nil, fmt.Errorf("worker is closed")
	}
	w.ids = dst[:0]
	err := w.engine.db.Scan(data, w.scratch, w.handler, nil)
	result := w.ids
	w.ids = nil
	return result, err
}
