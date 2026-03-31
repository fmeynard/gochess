package movegen

import (
	board "chessV2/internal/board"
	"math/bits"
)

type Position = board.Position
type Move = board.Move
type Piece = board.Piece

const (
	FenStartPos = board.FenStartPos

	NoPiece = board.NoPiece

	King   = board.King
	Queen  = board.Queen
	Pawn   = board.Pawn
	Knight = board.Knight
	Bishop = board.Bishop
	Rook   = board.Rook

	White = board.White
	Black = board.Black

	A1 = board.A1
	B1 = board.B1
	C1 = board.C1
	D1 = board.D1
	E1 = board.E1
	F1 = board.F1
	G1 = board.G1
	H1 = board.H1
	A2 = board.A2
	B2 = board.B2
	C2 = board.C2
	D2 = board.D2
	E2 = board.E2
	F2 = board.F2
	G2 = board.G2
	H2 = board.H2
	A3 = board.A3
	B3 = board.B3
	C3 = board.C3
	D3 = board.D3
	E3 = board.E3
	F3 = board.F3
	G3 = board.G3
	H3 = board.H3
	A4 = board.A4
	B4 = board.B4
	C4 = board.C4
	D4 = board.D4
	E4 = board.E4
	F4 = board.F4
	G4 = board.G4
	H4 = board.H4
	A5 = board.A5
	B5 = board.B5
	C5 = board.C5
	D5 = board.D5
	E5 = board.E5
	F5 = board.F5
	G5 = board.G5
	H5 = board.H5
	A6 = board.A6
	B6 = board.B6
	C6 = board.C6
	D6 = board.D6
	E6 = board.E6
	F6 = board.F6
	G6 = board.G6
	H6 = board.H6
	A7 = board.A7
	B7 = board.B7
	C7 = board.C7
	D7 = board.D7
	E7 = board.E7
	F7 = board.F7
	G7 = board.G7
	H7 = board.H7
	A8 = board.A8
	B8 = board.B8
	C8 = board.C8
	D8 = board.D8
	E8 = board.E8
	F8 = board.F8
	G8 = board.G8
	H8 = board.H8

	NoCastle        = board.NoCastle
	KingSideCastle  = board.KingSideCastle
	QueenSideCastle = board.QueenSideCastle
	NoEnPassant     = board.NoEnPassant

	NotCalculated = board.NotCalculated
	KingIsSafe    = board.KingIsSafe
	KingIsCheck   = board.KingIsCheck
)

func NewPositionFromFEN(fen string) (*Position, error) {
	return board.NewPositionFromFEN(fen)
}

var promotionFlags = [4]int8{
	board.QueenPromotion,
	board.KnightPromotion,
	board.BishopPromotion,
	board.RookPromotion,
}

func absInt8(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

func leastSignificantOne(bb uint64) int8 {
	return int8(bits.TrailingZeros64(bb))
}

func isOnBoard(file, rank int8) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

func isSameLineOrRow(start, end, direction int8) bool {
	switch direction {
	case 1, -1:
		return start/8 == end/8
	case 8, -8:
		return start%8 == end%8
	default:
		return absInt8(start%8-end%8) == absInt8(start/8-end/8)
	}
}

func isPromotionSquare(color, idx int8) bool {
	if color == White {
		return idx >= A8 && idx <= H8
	}

	return idx >= A1 && idx <= H1
}
