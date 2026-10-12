package corpus

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
}

type Record struct {
	Source string
	Index  int
	Data   []byte
}

func Normalize(r Request) []byte {
	// Reserve exactly the normalized size so the full corpus does not retain
	// spare capacity from repeated append growth for every request.
	size := len(r.Method) + 1 + len(r.URL) + 2 + 2 + len(r.Data)
	for k, v := range r.Headers {
		size += len(k) + 2 + len(v) + 2
	}
	b := make([]byte, 0, size)
	b = append(b, r.Method...)
	b = append(b, ' ')
	b = append(b, r.URL...)
	b = append(b, '\r', '\n')
	keys := make([]string, 0, len(r.Headers))
	for k := range r.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b = append(b, k+": "+r.Headers[k]+"\r\n"...)
	}
	b = append(b, "\r\n"...)
	return append(b, r.Data...)
}

type reader struct {
	root    string
	files   []string
	file    *os.File
	decoder *json.Decoder
	source  string
	index   int
}

func (r *reader) close() {
	if r.file != nil {
		_ = r.file.Close()
		r.file = nil
	}
}

func (r *reader) next() (Record, error) {
	for {
		if r.decoder == nil {
			if len(r.files) == 0 {
				return Record{}, io.EOF
			}
			path := r.files[0]
			r.files = r.files[1:]
			f, err := os.Open(path)
			if err != nil {
				return Record{}, err
			}
			r.file, r.decoder, r.index = f, json.NewDecoder(f), 0
			r.source, _ = filepath.Rel(r.root, path)
			r.source = filepath.ToSlash(r.source)
			tok, err := r.decoder.Token()
			if err != nil || tok != json.Delim('[') {
				return Record{}, fmt.Errorf("%s record 0: expected JSON array: %v", r.source, err)
			}
		}
		if !r.decoder.More() {
			tok, err := r.decoder.Token()
			if err != nil || tok != json.Delim(']') {
				return Record{}, fmt.Errorf("%s record %d: invalid array end: %v", r.source, r.index, err)
			}
			var extra json.RawMessage
			if err := r.decoder.Decode(&extra); err != io.EOF {
				return Record{}, fmt.Errorf("%s record %d: trailing JSON: %v", r.source, r.index, err)
			}
			r.close()
			r.decoder = nil
			continue
		}
		var raw map[string]json.RawMessage
		if err := r.decoder.Decode(&raw); err != nil {
			return Record{}, fmt.Errorf("%s record %d: %w", r.source, r.index, err)
		}
		var req Request
		for name, dst := range map[string]interface{}{"method": &req.Method, "url": &req.URL, "headers": &req.Headers, "data": &req.Data} {
			value, ok := raw[name]
			if !ok || string(value) == "null" {
				return Record{}, fmt.Errorf("%s record %d: missing or null %s", r.source, r.index, name)
			}
			if err := json.Unmarshal(value, dst); err != nil {
				return Record{}, fmt.Errorf("%s record %d field %s: %w", r.source, r.index, name, err)
			}
		}
		// json.Unmarshal permits null string values in maps; reject those explicitly.
		var headers map[string]json.RawMessage
		_ = json.Unmarshal(raw["headers"], &headers)
		for k, v := range headers {
			if string(v) == "null" {
				return Record{}, fmt.Errorf("%s record %d: null header %s", r.source, r.index, k)
			}
		}
		record := Record{Source: r.source, Index: r.index, Data: Normalize(req)}
		r.index++
		return record, nil
	}
}

func Load(root string, limit int) ([]Record, error) {
	if limit < 0 {
		return nil, fmt.Errorf("limit must be nonnegative")
	}
	readers := make([]*reader, 0, 2)
	for _, folder := range []string{"legitimate", "malicious"} {
		path := filepath.Join(root, folder)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return nil, fmt.Errorf("missing corpus directory %s", path)
		}
		files, err := filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return nil, err
		}
		sort.Strings(files)
		if len(files) == 0 {
			return nil, fmt.Errorf("no JSON files in %s", path)
		}
		r := &reader{root: root, files: files}
		readers = append(readers, r)
		defer r.close()
	}
	var result []Record
	done := [2]bool{}
	for !done[0] || !done[1] {
		for i, r := range readers {
			if done[i] {
				continue
			}
			record, err := r.next()
			if err == io.EOF {
				done[i] = true
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, record)
			if limit > 0 && len(result) == limit {
				return result, nil
			}
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("empty corpus")
	}
	return result, nil
}
