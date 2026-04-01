package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCentipawnScore(t *testing.T) {
	score, ok := parseCentipawnScore("info depth 12 score cp -84 pv e2e4")
	assert.True(t, ok)
	assert.Equal(t, -84, score)
}

func TestParseMateScore(t *testing.T) {
	score, ok := parseCentipawnScore("info depth 18 score mate 3 pv e2e4")
	assert.True(t, ok)
	assert.Equal(t, 30000, score)

	score, ok = parseCentipawnScore("info depth 18 score mate -2 pv e7e5")
	assert.True(t, ok)
	assert.Equal(t, -30000, score)
}
