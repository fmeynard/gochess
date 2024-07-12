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
		{"a1", A1},
		{"h1", H1},
		{"a8", A8},
		{"h8", H8},
		{"d4", D4},
		{"e5", E5},
	}

	for _, d := range data {
		assert.Equal(t, d.expectedIdx, SquareToIdx(d.strIdentifier))
	}

	assert.Panics(t, func() { SquareToIdx("") })
	assert.Panics(t, func() { SquareToIdx("b11") })
	assert.Panics(t, func() { SquareToIdx("c9") })
	assert.Panics(t, func() { SquareToIdx("-1") })
}

func Test_IdxToSquare(t *testing.T) {
	data := []struct {
		strIdentifier string
		idx           int8
	}{
		{"a1", A1},
		{"h1", H1},
		{"a8", A8},
		{"h8", H8},
		{"d4", D4},
		{"e5", E5},
	}

	for _, d := range data {
		assert.Equal(t, d.strIdentifier, IdxToSquare(d.idx))
	}

	assert.Panics(t, func() { IdxToSquare(-1) })
	assert.Panics(t, func() { IdxToSquare(64) })
}

func Test_RankFromIdx(t *testing.T) {
	data := []struct {
		idx  int8
		rank int8
	}{
		{A1, 0},
		{B2, 1},
		{C3, 2},
		{A4, 3},
		{H5, 4},
	}

	for _, d := range data {
		assert.Equal(t, d.rank, RankFromIdx(d.idx))
	}
}

func Test_FileFromIdx(t *testing.T) {
	data := []struct {
		idx  int8
		file int8
	}{
		{A1, 0},
		{B1, 1},
		{C1, 2},
		{D1, 3},
		{E1, 4},
		{H8, 7},
	}

	for _, d := range data {
		assert.Equal(t, d.file, FileFromIdx(d.idx))
	}
}

func Test_RankAndFile(t *testing.T) {
	data := []struct {
		idx  int8
		rank int8
		file int8
	}{
		{A1, 0, 0},
		{B2, 1, 1},
		{C3, 2, 2},
		{A8, 7, 0},
		{H1, 0, 7},
	}

	for _, d := range data {
		rank, file := RankAndFile(d.idx)
		assert.Equal(t, d.rank, rank)
		assert.Equal(t, d.file, file)
	}
}
