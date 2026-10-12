package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andreyfesunov/waap_regex_bench/internal/baseline"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
)

func getSignatures(dataset string) ([]signatures.Signature, error) {
	f, err := os.Open(filepath.Join(dataset, "nmap-service-probes"))
	if err != nil {
		return nil, err
	}
	sigs, parseErr := signatures.Parse(f)
	closeErr := f.Close()
	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return sigs, nil
}

func getBaselineEngine(sigs []signatures.Signature) (*baseline.Engine, []signatures.Signature, []baseline.Rejection, error) {
	engine, accepted, rejected := baseline.Compile(sigs)
	encoder := json.NewEncoder(os.Stderr)
	for _, rejection := range rejected {
		if err := encoder.Encode(rejection); err != nil {
			return nil, nil, rejected, err
		}
	}
	if len(accepted) == 0 {
		return nil, nil, rejected, fmt.Errorf("no compatible regex patterns")
	}
	return engine, accepted, rejected, nil
}
