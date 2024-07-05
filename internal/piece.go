package internal

type Piece int8
type PieceColor int8

// 11000
const (
	NoPiece Piece = 0
	King    int8  = 1
	Queen   int8  = 2
	Pawn    int8  = 3
	Knight  int8  = 4
	Bishop  int8  = 5
	Rook    int8  = 6

	White int8 = 8
	Black int8 = 16
)

func (p Piece) IsWhite() bool {
	return p&(1<<3) != 0
}

func (p Piece) Color() int8 {
	if p.IsWhite() {
		return White
	}

	return Black
}

func (p Piece) IsType(pieceType int8) bool {
	// & 7 : Mask out color bits
	return (int8(p) & 7) == int8(pieceType)
}
