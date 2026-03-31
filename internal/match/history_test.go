package match

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSummaryMarkdownRow(t *testing.T) {
	summary := Summary{
		Date:          time.Date(2026, 3, 31, 22, 0, 0, 0, time.UTC),
		Current:       "691a34a",
		Opponent:      "self@691a34a",
		MoveTime:      50 * time.Millisecond,
		Games:         2,
		CurrentWins:   0,
		Draws:         2,
		CurrentLosses: 0,
		Notes:         "sample smoke run",
	}

	assert.Equal(
		t,
		"| 2026-03-31T22:00:00Z | 691a34a | self@691a34a | 50ms | 2 | 1.0/2 | 0/2/0 | sample smoke run |\n",
		summary.MarkdownRow(),
	)
}
