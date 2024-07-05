package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SquareToIdx(t *testing.T) {
	data := []struct {
		strIdentifier string
		expectedIdx   int8
	}{
		{"a1", 0},
		{"h1", 7},
		{"a8", 56},
		{"h8", 63},
		{"d4", 27},
		{"e5", 36},
	}

	for _, d := range data {
		assert.Equal(t, d.expectedIdx, SquareToIdx(d.strIdentifier))
	}
}

func Test_IdxToSquare(t *testing.T) {
	data := []struct {
		idx           int8
		strIdentifier string
	}{
		{0, "a1"},
		{7, "h1"},
		{8, "a2"},
		{27, "d4"},
		{36, "e5"},
		{63, "h8"},
	}

	for _, d := range data {
		assert.Equal(t, d.strIdentifier, IdxToSquare(d.idx))
	}
}
