package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCentipawnScore(t *testing.T) {
	score, ok := parseCentipawnScore("info depth 12 score cp -84 pv e2e4")
	assert.True(t, ok)
	assert.Equal(t, -84, score.CP)
	assert.False(t, score.IsMate)
}

func TestParseMateScore(t *testing.T) {
	score, ok := parseCentipawnScore("info depth 18 score mate 3 pv e2e4")
	assert.True(t, ok)
	assert.Equal(t, 30000, score.CP)
	assert.True(t, score.IsMate)

	score, ok = parseCentipawnScore("info depth 18 score mate -2 pv e7e5")
	assert.True(t, ok)
	assert.Equal(t, -30000, score.CP)
	assert.True(t, score.IsMate)
}

func TestReadLinesHandlesLongInput(t *testing.T) {
	out := make(chan string, 1)
	errs := make(chan error, 1)
	longLine := strings.Repeat("x", 200000) + "\n"

	go readLines(strings.NewReader(longLine), out, errs)

	line, ok := <-out
	assert.True(t, ok)
	assert.Len(t, line, 200000)

	_, stillOpen := <-out
	assert.False(t, stillOpen)

	select {
	case err := <-errs:
		assert.NoError(t, err)
	default:
	}
}

func TestNonKingMaterial(t *testing.T) {
	material, err := nonKingMaterial("4k3/8/8/8/8/8/4Q3/4K3 w - - 0 1")
	assert.NoError(t, err)
	assert.Equal(t, 900, material)

	material, err = nonKingMaterial("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
	assert.NoError(t, err)
	assert.Equal(t, 0, material)
}
