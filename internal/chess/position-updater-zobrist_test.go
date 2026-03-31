package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainPositionUpdater_DoesNotTrackZobrist(t *testing.T) {
	pos, err := NewPositionFromFEN(FenStartPos)
	if err != nil {
		t.Fatal(err)
	}

	updater := NewPlainPositionUpdater(NewPseudoLegalMoveGenerator())
	initialKey := pos.zobristKey
	move := NewMove(Piece(White|Pawn), E2, E4, PawnDoubleMove)
	history := updater.MakeMove(pos, move)

	assert.Equal(t, uint64(0), history.zobristKey)
	assert.Equal(t, initialKey, pos.zobristKey)

	updater.UnMakeMove(pos, history)
	assert.Equal(t, initialKey, pos.zobristKey)
}

func TestZobristPositionUpdater_RestoresKeyOnUnmake(t *testing.T) {
	pos, err := NewPositionFromFEN(FenStartPos)
	if err != nil {
		t.Fatal(err)
	}

	updater := NewPositionUpdater(NewPseudoLegalMoveGenerator())
	initialKey := pos.zobristKey
	move := NewMove(Piece(White|Pawn), E2, E4, PawnDoubleMove)
	history := updater.MakeMove(pos, move)

	assert.NotEqual(t, initialKey, pos.zobristKey)
	assert.Equal(t, initialKey, history.zobristKey)

	updater.UnMakeMove(pos, history)
	assert.Equal(t, initialKey, pos.zobristKey)
}
