//go:build hyperscan && cgo

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andreyfesunov/waap_regex_bench/internal/bench"
	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	hs "github.com/andreyfesunov/waap_regex_bench/internal/hyperscan"
	"os"
	"runtime"
	"time"
)

func run() (err error) {
	config, err := getConfig()
	if err != nil {
		return err
	}
	start := time.Now()
	sigs, hash, err := getSignatures(config.Dataset)
	if err != nil {
		return err
	}
	total := len(sigs)
	compileStart := time.Now()
	engine, accepted, rejected, excluded, err := getHyperscanEngine(sigs, config.PatternSet)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, engine.Close()) }()
	compileSeconds := time.Since(compileStart).Seconds()
	manifest := config.Output + ".patterns.json"
	if err := saveManifest(manifest, hash, config.PatternSet, accepted); err != nil {
		return err
	}
	dbSize, err := engine.Size()
	if err != nil {
		return err
	}
	probe, err := engine.NewWorker()
	if err != nil {
		return err
	}
	scratchSize, sizeErr := probe.Size()
	closeErr := probe.Close()
	if err := errors.Join(sizeErr, closeErr); err != nil {
		return err
	}
	records, err := corpus.Load(config.Dataset, config.Limit)
	if err != nil {
		return err
	}
	prepareSeconds := time.Since(start).Seconds()
	fmt.Fprintf(os.Stderr, "prepared %d requests, %d/%d patterns; starting %d worker(s)\n", len(records), len(accepted), total, config.Workers)
	results, stats, err := bench.Run(records, accepted, config.Workers, func() (bench.Worker, error) {
		w, err := engine.NewWorker()
		if err != nil {
			return nil, err
		}
		return w, nil
	})
	if err != nil {
		return err
	}
	writeStart := time.Now()
	if err := bench.WriteJSONL(config.Output, results); err != nil {
		return err
	}
	summary := struct {
		bench.Stats
		Engine           string  `json:"engine"`
		HyperscanVersion string  `json:"hyperscan_version"`
		Mode             string  `json:"mode"`
		PatternSet       string  `json:"pattern_set"`
		PatternsTotal    int     `json:"patterns_total"`
		PatternsExcluded int     `json:"patterns_excluded"`
		PatternsRejected int     `json:"patterns_rejected"`
		PatternsAccepted int     `json:"patterns_accepted"`
		DatabaseBytes    int     `json:"database_bytes"`
		ScratchBytes     int     `json:"scratch_bytes_per_worker"`
		CompileSeconds   float64 `json:"compile_seconds"`
		PrepareSeconds   float64 `json:"prepare_seconds"`
		WriteSeconds     float64 `json:"write_seconds"`
		WallSeconds      float64 `json:"wall_seconds"`
		GoVersion        string  `json:"go_version"`
		GOOS             string  `json:"goos"`
		GOARCH           string  `json:"goarch"`
		AvailableCPUs    int     `json:"available_cpus"`
		SourceSHA256     string  `json:"source_sha256"`
		PatternManifest  string  `json:"pattern_manifest"`
		Output           string  `json:"output"`
	}{stats, "hyperscan", hs.Version(), "bytes", config.PatternSet, total, excluded, len(rejected), len(accepted), dbSize, scratchSize, compileSeconds, prepareSeconds, time.Since(writeStart).Seconds(), time.Since(start).Seconds(), runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), hash, manifest, config.Output}
	return json.NewEncoder(os.Stdout).Encode(summary)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
