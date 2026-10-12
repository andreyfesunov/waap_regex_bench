package main

import (
	"flag"
	"fmt"
)

type HyperscanConfig struct {
	Dataset    string
	Workers    int
	Limit      int
	Output     string
	PatternSet string
}

func getConfig() (HyperscanConfig, error) {
	var c HyperscanConfig
	flag.StringVar(&c.Dataset, "dataset", "dataset", "dataset root")
	flag.IntVar(&c.Workers, "workers", 1, "workers and GOMAXPROCS")
	flag.IntVar(&c.Limit, "limit", 100000, "request limit; 0 loads full corpus")
	flag.StringVar(&c.Output, "output", "hyperscan-results.jsonl", "results JSONL (overwritten)")
	flag.StringVar(&c.PatternSet, "pattern-set", "all", "all or common with Go regexp")
	flag.Parse()
	if flag.NArg() != 0 {
		return c, fmt.Errorf("positional arguments are not supported")
	}
	return c, validateConfig(c)
}
func validateConfig(c HyperscanConfig) error {
	if c.Dataset == "" || c.Output == "" {
		return fmt.Errorf("dataset and output must not be empty")
	}
	if c.Workers < 1 || c.Limit < 0 {
		return fmt.Errorf("require workers > 0 and limit >= 0")
	}
	if c.PatternSet != "all" && c.PatternSet != "common" {
		return fmt.Errorf("pattern-set must be all or common")
	}
	return nil
}
