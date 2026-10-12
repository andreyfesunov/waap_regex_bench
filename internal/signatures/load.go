package signatures

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Load(dataset string) ([]Signature, string, error) {
	f, err := os.Open(filepath.Join(dataset, "nmap-service-probes"))
	if err != nil {
		return nil, "", err
	}
	h := sha256.New()
	sigs, parseErr := Parse(io.TeeReader(f, h))
	closeErr := f.Close()
	if parseErr != nil {
		return nil, "", parseErr
	}
	if closeErr != nil {
		return nil, "", closeErr
	}
	return sigs, hex.EncodeToString(h.Sum(nil)), nil
}

type Manifest struct {
	SourceSHA256 string `json:"source_sha256"`
	PatternSet   string `json:"pattern_set"`
	Mode         string `json:"mode"`
	IDs          []int  `json:"ids"`
}

func SaveManifest(path, hash, set string, sigs []Signature) error {
	m := Manifest{SourceSHA256: hash, PatternSet: set, Mode: "bytes", IDs: make([]int, 0, len(sigs))}
	for _, s := range sigs {
		m.IDs = append(m.IDs, s.ID)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func SelectManifest(path, hash string, sigs []Signature) ([]Signature, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	if m.SourceSHA256 != hash {
		return nil, fmt.Errorf("manifest source SHA-256 does not match nmap-service-probes")
	}
	if len(m.IDs) == 0 {
		return nil, fmt.Errorf("manifest has no patterns")
	}
	catalog := make(map[int]Signature, len(sigs))
	for _, s := range sigs {
		catalog[s.ID] = s
	}
	seen := make(map[int]bool, len(m.IDs))
	selected := make([]Signature, 0, len(m.IDs))
	for _, id := range m.IDs {
		if seen[id] {
			return nil, fmt.Errorf("duplicate manifest ID %d", id)
		}
		s, ok := catalog[id]
		if !ok {
			return nil, fmt.Errorf("unknown manifest ID %d", id)
		}
		seen[id] = true
		selected = append(selected, s)
	}
	return selected, nil
}
