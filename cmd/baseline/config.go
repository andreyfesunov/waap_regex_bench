package main

import (
	"flag"
	"fmt"
)

type BaselineConfig struct {
	Dataset    string
	Workers    int
	Limit      int
	Output     string
	PatternIDs string
}

func getConfig() (BaselineConfig, error) {
	var config BaselineConfig
	flag.StringVar(&config.Dataset, "dataset", "dataset", "dataset root")
	flag.IntVar(&config.Workers, "workers", 1, "number of workers and GOMAXPROCS")
	flag.IntVar(&config.Limit, "limit", 100000, "request limit; 0 loads the full corpus")
	flag.StringVar(&config.Output, "output", "baseline-results.jsonl", "results JSONL (overwritten)")
	flag.StringVar(&config.PatternIDs, "pattern-ids", "", "Hyperscan pattern manifest for identical pattern selection")
	flag.Parse()
	if flag.NArg() != 0 {
		return config, fmt.Errorf("positional arguments are not supported")
	}
	return config, validateConfig(config)
}

func validateConfig(config BaselineConfig) error {
	if config.Dataset == "" {
		return fmt.Errorf("dataset must not be empty")
	}
	if config.Workers < 1 {
		return fmt.Errorf("workers must be positive")
	}
	if config.Limit < 0 {
		return fmt.Errorf("limit must be nonnegative")
	}
	if config.Output == "" {
		return fmt.Errorf("output must not be empty")
	}
	return nil
}
