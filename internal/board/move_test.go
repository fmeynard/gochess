package board

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoveHelpers_ClassifyMove(t *testing.T) {
	data := map[string]struct {
		fen      string
		piece    Piece
		startIdx int8
		endIdx   int8
		expected int8
	}{
		"Normal move": {
			fen:      FenStartPos,
			piece:    Piece(White | Knight),
			startIdx: G1,
			endIdx:   F3,
			expected: NormalMove,
		},
		"Capture move": {
			fen:      "8/8/8/8/3p4/8/3R4/8 w - - 0 1",
			piece:    Piece(White | Rook),
			startIdx: D2,
			endIdx:   D4,
			expected: Capture,
		},
		"Castle move": {
			fen:      "8/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			piece:    Piece(White | King),
			startIdx: E1,
			endIdx:   G1,
			expected: Castle,
		},
		"Pawn double move": {
			fen:      FenStartPos,
			piece:    Piece(White | Pawn),
			startIdx: E2,
			endIdx:   E4,
			expected: PawnDoubleMove,
		},
		"En passant move": {
			fen:      "8/8/8/3pP3/8/8/8/8 w - d6 0 1",
			piece:    Piece(White | Pawn),
			startIdx: E5,
			endIdx:   D6,
			expected: EnPassant,
		},
	}

	for name, d := range data {
		t.Run(name, func(t *testing.T) {
			pos, err := NewPositionFromFEN(d.fen)
			if err != nil {
				t.Fatal(err)
			}

			actual := int8(NormalMove)
			if d.piece.Type() == King && absInt8(d.endIdx-d.startIdx) == 2 {
				actual = Castle
			} else if d.piece.Type() == Pawn && pos.enPassantIdx != NoEnPassant && d.endIdx == pos.enPassantIdx {
				actual = EnPassant
			} else if d.piece.Type() == Pawn && absInt8(d.endIdx-d.startIdx) == 16 {
				actual = PawnDoubleMove
			} else if pos.board[d.endIdx] != NoPiece {
				actual = Capture
			}

			assert.Equal(t, d.expected, actual)
		})
	}
}

func TestMoveHelpers_IsCastleAndEnPassant(t *testing.T) {
	pos, err := NewPositionFromFEN("8/8/8/3pP3/8/8/8/8 w - d6 0 1")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, isCastleMove(NewMove(Piece(White|King), E1, G1, Castle)))
	assert.True(t, isCastleMove(NewMove(Piece(White|King), E1, G1, NormalMove)))
	assert.False(t, isCastleMove(NewMove(Piece(White|King), E1, F1, NormalMove)))

	move := NewMove(Piece(White|Pawn), E5, D6, EnPassant)
	assert.True(t, move.flag == EnPassant || (move.piece.Type() == Pawn && pos.enPassantIdx != NoEnPassant && move.EndIdx() == pos.enPassantIdx && pos.PieceAt(move.EndIdx()) == NoPiece && absInt8(FileFromIdx(move.EndIdx())-FileFromIdx(move.StartIdx())) == 1))

	move = NewMove(Piece(White|Pawn), E5, D6, NormalMove)
	assert.True(t, move.flag == EnPassant || (move.piece.Type() == Pawn && pos.enPassantIdx != NoEnPassant && move.EndIdx() == pos.enPassantIdx && pos.PieceAt(move.EndIdx()) == NoPiece && absInt8(FileFromIdx(move.EndIdx())-FileFromIdx(move.StartIdx())) == 1))

	move = NewMove(Piece(White|Pawn), E5, E6, NormalMove)
	assert.False(t, move.flag == EnPassant || (move.piece.Type() == Pawn && pos.enPassantIdx != NoEnPassant && move.EndIdx() == pos.enPassantIdx && pos.PieceAt(move.EndIdx()) == NoPiece && absInt8(FileFromIdx(move.EndIdx())-FileFromIdx(move.StartIdx())) == 1))
}
