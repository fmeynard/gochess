package match

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type RecordWriter struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
}

func NewRecordWriter(path string) (*RecordWriter, error) {
	if path == "" {
		return nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &RecordWriter{
		file: file,
		enc:  json.NewEncoder(file),
	}, nil
}

func (w *RecordWriter) Write(record MoveRecord) error {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	return w.enc.Encode(record)
}

func (w *RecordWriter) Close() error {
	if w == nil {
		return nil
	}
	return w.file.Close()
}
