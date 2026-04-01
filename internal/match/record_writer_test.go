package match

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordWriterWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "records.jsonl")

	writer, err := NewRecordWriter(path)
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	record := MoveRecord{
		GameIndex:      1,
		Ply:            2,
		CurrentAsWhite: true,
		SideToMove:     "black",
		Player:         "current",
		Move:           "e2e4",
		FENBefore:      "before fen",
		FENAfter:       "after fen",
		Timestamp:      time.Date(2026, 4, 1, 19, 0, 0, 0, time.UTC),
	}
	assert.NoError(t, writer.Write(record))
	assert.NoError(t, writer.Close())

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	var decoded MoveRecord
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, record.GameIndex, decoded.GameIndex)
	assert.Equal(t, record.Ply, decoded.Ply)
	assert.Equal(t, record.Move, decoded.Move)
	assert.Equal(t, record.FENBefore, decoded.FENBefore)
	assert.Equal(t, record.FENAfter, decoded.FENAfter)
}
