//go:build hyperscan && cgo

package main

import (
	"encoding/json"
	"github.com/andreyfesunov/waap_regex_bench/internal/baseline"
	hs "github.com/andreyfesunov/waap_regex_bench/internal/hyperscan"
	"github.com/andreyfesunov/waap_regex_bench/internal/signatures"
	"os"
)

func getSignatures(dataset string) ([]signatures.Signature, string, error) {
	return signatures.Load(dataset)
}
func saveManifest(path, hash, set string, sigs []signatures.Signature) error {
	return signatures.SaveManifest(path, hash, set, sigs)
}
func getHyperscanEngine(sigs []signatures.Signature, set string) (*hs.Engine, []signatures.Signature, []hs.Rejection, int, error) {
	excluded := 0
	if set == "common" {
		_, accepted, rejected := baseline.Compile(sigs)
		excluded = len(rejected)
		sigs = accepted
		for _, r := range rejected {
			if err := json.NewEncoder(os.Stderr).Encode(struct {
				baseline.Rejection
				Stage string `json:"stage"`
			}{r, "go-selection"}); err != nil {
				return nil, nil, nil, excluded, err
			}
		}
	}
	e, accepted, rejected, err := hs.Compile(sigs)
	for _, r := range rejected {
		if writeErr := json.NewEncoder(os.Stderr).Encode(r); writeErr != nil {
			if e != nil {
				_ = e.Close()
			}
			return nil, nil, rejected, excluded, writeErr
		}
	}
	return e, accepted, rejected, excluded, err
}
