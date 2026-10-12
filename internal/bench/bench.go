package bench

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
)

type Worker interface {
	Match([]byte, []int) ([]int, error)
}
type Factory func() (Worker, error)

type Service struct {
	Name      string `json:"service"`
	Kind      string `json:"kind"`
	Probe     string `json:"probe"`
	PatternID int    `json:"pattern_id"`
}
type Result struct {
	Source   string    `json:"source"`
	Index    int       `json:"index"`
	Services []Service `json:"services"`
}
type Stats struct {
	Requests          int     `json:"requests"`
	Bytes             int64   `json:"bytes"`
	Workers           int     `json:"workers"`
	GOMAXPROCS        int     `json:"gomaxprocs"`
	PatternMatches    int64   `json:"pattern_matches"`
	ServiceMatches    int64   `json:"service_matches"`
	MatchSeconds      float64 `json:"match_seconds"`
	RequestsPerSecond float64 `json:"requests_per_second"`
}

func better(a, b signatures.Signature) bool {
	if a.Kind != b.Kind {
		return a.Kind == "match"
	}
	if a.ProbeOrder != b.ProbeOrder {
		return a.ProbeOrder < b.ProbeOrder
	}
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.ID < b.ID
}

func resolve(ids []int, catalog map[int]signatures.Signature, chosen map[string]signatures.Signature) ([]Service, error) {
	clear(chosen)
	for _, id := range ids {
		s, ok := catalog[id]
		if !ok {
			return nil, fmt.Errorf("engine returned unknown pattern ID %d", id)
		}
		old, exists := chosen[s.Service]
		if !exists || better(s, old) {
			chosen[s.Service] = s
		}
	}
	services := make([]Service, 0, len(chosen))
	for _, s := range chosen {
		services = append(services, Service{s.Service, s.Kind, s.Probe, s.ID})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}

func Run(records []corpus.Record, sigs []signatures.Signature, workers int, factory Factory) (results []Result, stats Stats, err error) {
	if workers < 1 {
		return nil, Stats{}, fmt.Errorf("workers must be positive")
	}
	catalog := make(map[int]signatures.Signature, len(sigs))
	for _, s := range sigs {
		if _, ok := catalog[s.ID]; ok {
			return nil, Stats{}, fmt.Errorf("duplicate pattern ID %d", s.ID)
		}
		catalog[s.ID] = s
	}
	instances := make([]Worker, workers)
	defer func() {
		for _, worker := range instances {
			if closer, ok := worker.(io.Closer); ok {
				err = errors.Join(err, closer.Close())
			}
		}
	}()
	for i := range instances {
		w, err := factory()
		instances[i] = w
		if err != nil {
			return nil, Stats{}, err
		}
		if w == nil {
			return nil, Stats{}, fmt.Errorf("factory returned nil worker")
		}
	}
	previous := runtime.GOMAXPROCS(workers)
	defer runtime.GOMAXPROCS(previous)
	results = make([]Result, len(records))
	stats = Stats{Requests: len(records), Workers: workers, GOMAXPROCS: runtime.GOMAXPROCS(0)}
	for _, r := range records {
		stats.Bytes += int64(len(r.Data))
	}
	var next atomic.Int64
	var stopped atomic.Bool
	var wg, ready sync.WaitGroup
	var once sync.Once
	var runErr error
	counts := make([]int64, workers)
	start := make(chan struct{})
	wg.Add(workers)
	ready.Add(workers)
	for i, w := range instances {
		go func(i int, w Worker) {
			defer wg.Done()
			ids := make([]int, 0, 16)
			chosen := make(map[string]signatures.Signature)
			ready.Done()
			<-start
			for !stopped.Load() {
				begin := int(next.Add(256) - 256)
				if begin >= len(records) {
					return
				}
				end := min(begin+256, len(records))
				for j := begin; j < end; j++ {
					var err error
					ids, err = w.Match(records[j].Data, ids)
					var services []Service
					if err == nil {
						services, err = resolve(ids, catalog, chosen)
					}
					if err != nil {
						once.Do(func() { runErr = fmt.Errorf("request %d: %w", j, err) })
						stopped.Store(true)
						return
					}
					counts[i] += int64(len(ids))
					results[j] = Result{records[j].Source, records[j].Index, services}
				}
			}
		}(i, w)
	}
	ready.Wait()
	t0 := time.Now()
	close(start)
	wg.Wait()
	stats.MatchSeconds = time.Since(t0).Seconds()
	for _, count := range counts {
		stats.PatternMatches += count
	}
	for _, result := range results {
		stats.ServiceMatches += int64(len(result.Services))
	}
	if stats.MatchSeconds > 0 {
		stats.RequestsPerSecond = float64(len(records)) / stats.MatchSeconds
	}
	return results, stats, runErr
}

func WriteJSONL(path string, results []Result) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	b := bufio.NewWriterSize(f, 1024*1024)
	encoder := json.NewEncoder(b)
	for _, r := range results {
		if err := encoder.Encode(r); err != nil {
			return err
		}
	}
	return b.Flush()
}
