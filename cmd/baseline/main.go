package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/andreyfesunov/waap_regex_bench/internal/bench"
	"github.com/andreyfesunov/waap_regex_bench/internal/corpus"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
)

func run() error {
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
	if config.PatternIDs != "" {
		sigs, err = signatures.SelectManifest(config.PatternIDs, hash, sigs)
		if err != nil {
			return err
		}
	}
	compileStart := time.Now()
	engine, accepted, rejected, err := getBaselineEngine(sigs)
	if err != nil {
		return err
	}

	compileSeconds := time.Since(compileStart).Seconds()
	if config.PatternIDs != "" && len(rejected) > 0 {
		return fmt.Errorf("manifest includes %d patterns incompatible with Go; refusing to change selection", len(rejected))
	}
	records, err := corpus.Load(config.Dataset, config.Limit)
	if err != nil {
		return err
	}

	prepareSeconds := time.Since(start).Seconds()
	fmt.Fprintf(os.Stderr, "prepared %d requests, %d/%d patterns; starting %d worker(s)\n", len(records), len(accepted), len(sigs), config.Workers)

	results, stats, err := bench.Run(records, accepted, config.Workers, func() (bench.Worker, error) { return engine.NewWorker(), nil })
	if err != nil {
		return err
	}

	writeStart := time.Now()

	if err := bench.WriteJSONL(config.Output, results); err != nil {
		return err
	}

	summary := struct {
		bench.Stats
		PrepareSeconds   float64 `json:"prepare_seconds"`
		WriteSeconds     float64 `json:"write_seconds"`
		WallSeconds      float64 `json:"wall_seconds"`
		PatternsTotal    int     `json:"patterns_total"`
		PatternsAccepted int     `json:"patterns_accepted"`
		PatternsRejected int     `json:"patterns_rejected"`
		GoVersion        string  `json:"go_version"`
		GOOS             string  `json:"goos"`
		GOARCH           string  `json:"goarch"`
		Output           string  `json:"output"`
		Engine           string  `json:"engine"`
		PatternsExcluded int     `json:"patterns_excluded"`
		CompileSeconds   float64 `json:"compile_seconds"`
		SourceSHA256     string  `json:"source_sha256"`
		PatternManifest  string  `json:"pattern_manifest,omitempty"`
	}{stats, prepareSeconds, time.Since(writeStart).Seconds(), time.Since(start).Seconds(), total, len(accepted), len(rejected), runtime.Version(), runtime.GOOS, runtime.GOARCH, config.Output, "go-regexp", total - len(sigs), compileSeconds, hash, config.PatternIDs}

	return json.NewEncoder(os.Stdout).Encode(summary)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
