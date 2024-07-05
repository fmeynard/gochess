package internal

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestPiece_IsWhite(t *testing.T) {
	assert.Equal(t, true, Piece(White|Pawn).IsWhite())
}

func TestPiece_IsType(t *testing.T) {
	assert.Equal(t, true, Piece(White|Queen).IsType(Queen))
	assert.Equal(t, true, Piece(Black|Queen).IsType(Queen))
	assert.Equal(t, false, Piece(White|Pawn).IsType(Queen))
	assert.Equal(t, false, Piece(Black|Rook).IsType(Queen))
}
