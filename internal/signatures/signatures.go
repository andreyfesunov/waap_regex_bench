package signatures

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Signature struct {
	ID         int    `json:"id"`
	Service    string `json:"service"`
	Kind       string `json:"kind"`
	Probe      string `json:"probe"`
	ProbeOrder int    `json:"probe_order"`
	Line       int    `json:"line"`
	Expression string `json:"expression"`
	Flags      string `json:"flags"`
}

func Parse(r io.Reader) ([]Signature, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	var result []Signature
	probe, order := "", -1
	line := 0
	for scanner.Scan() {
		line++
		s := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(s)
		if len(fields) == 0 || strings.HasPrefix(s, "#") {
			continue
		}
		if fields[0] == "Probe" {
			if len(fields) < 4 {
				return nil, fmt.Errorf("line %d: malformed Probe", line)
			}
			probe, order = fields[2], order+1
			continue
		}
		if fields[0] != "match" && fields[0] != "softmatch" {
			continue
		}
		if order < 0 {
			return nil, fmt.Errorf("line %d: malformed signature or missing Probe", line)
		}
		sig, err := parseSignatureLine(s)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		sig.ID = len(result)
		sig.Probe = probe
		sig.ProbeOrder = order
		sig.Line = line
		result = append(result, sig)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("line %d: %w", line, err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no signatures found")
	}
	return result, nil
}

func parseSignatureLine(line string) (Signature, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 || (fields[0] != "match" && fields[0] != "softmatch") {
		return Signature{}, fmt.Errorf("malformed signature or missing Probe")
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
	rest = strings.TrimSpace(strings.TrimPrefix(rest, fields[1]))
	expression, flags, err := parseRegex(rest)
	if err != nil {
		return Signature{}, err
	}
	return Signature{
		Service:    fields[1],
		Kind:       fields[0],
		Expression: expression,
		Flags:      flags,
	}, nil
}

func parseRegex(text string) (string, string, error) {
	if len(text) < 3 || text[0] != 'm' {
		return "", "", fmt.Errorf("missing regex delimiter")
	}
	delim, end := text[1], -1
	for i := 2; i < len(text); i++ {
		if text[i] == '\\' {
			i++
			continue
		}
		if text[i] == delim {
			end = i
			break
		}
	}
	if end < 0 {
		return "", "", fmt.Errorf("unterminated regex")
	}
	flagEnd := end + 1
	for flagEnd < len(text) && text[flagEnd] != ' ' && text[flagEnd] != '\t' {
		flagEnd++
	}
	// An escaped delimiter denotes a literal delimiter. Preserve regex escapes.
	expr := text[2:end]
	var b strings.Builder
	for i := 0; i < len(expr); i++ {
		if expr[i] == '\\' && i+1 < len(expr) {
			if expr[i+1] == delim && !strings.ContainsRune(`\.+*?()|[]{}^$`, rune(delim)) {
				b.WriteByte(delim)
			} else {
				b.WriteByte(expr[i])
				b.WriteByte(expr[i+1])
			}
			i++
		} else {
			b.WriteByte(expr[i])
		}
	}
	return b.String(), text[end+1 : flagEnd], nil
}
