package internal

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestPiece_IsWhite(t *testing.T) {
	assert.Equal(t, true, Piece(White|Pawn).IsWhite())
}

func TestPiece_IsType(t *testing.T) {
	assert.Equal(t, Queen, Piece(White|Queen).Type())
	assert.Equal(t, Queen, Piece(Black|Queen).Type())
	assert.Equal(t, Pawn, Piece(White|Pawn).Type())
	assert.Equal(t, Rook, Piece(Black|Rook).Type())
}
