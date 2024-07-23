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

const (
	A1 int8 = 0
	B1 int8 = 1
	C1 int8 = 2
	D1 int8 = 3
	E1 int8 = 4
	F1 int8 = 5
	G1 int8 = 6
	H1 int8 = 7
	A2 int8 = 8
	B2 int8 = 9
	C2 int8 = 10
	D2 int8 = 11
	E2 int8 = 12
	F2 int8 = 13
	G2 int8 = 14
	H2 int8 = 15
	A3 int8 = 16
	B3 int8 = 17
	C3 int8 = 18
	D3 int8 = 19
	E3 int8 = 20
	F3 int8 = 21
	G3 int8 = 22
	H3 int8 = 23
	A4 int8 = 24
	B4 int8 = 25
	C4 int8 = 26
	D4 int8 = 27
	E4 int8 = 28
	F4 int8 = 29
	G4 int8 = 30
	H4 int8 = 31
	A5 int8 = 32
	B5 int8 = 33
	C5 int8 = 34
	D5 int8 = 35
	E5 int8 = 36
	F5 int8 = 37
	G5 int8 = 38
	H5 int8 = 39
	A6 int8 = 40
	B6 int8 = 41
	C6 int8 = 42
	D6 int8 = 43
	E6 int8 = 44
	F6 int8 = 45
	G6 int8 = 46
	H6 int8 = 47
	A7 int8 = 48
	B7 int8 = 49
	C7 int8 = 50
	D7 int8 = 51
	E7 int8 = 52
	F7 int8 = 53
	G7 int8 = 54
	H7 int8 = 55
	A8 int8 = 56
	B8 int8 = 57
	C8 int8 = 58
	D8 int8 = 59
	E8 int8 = 60
	F8 int8 = 61
	G8 int8 = 62
	H8 int8 = 63
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

func (p Piece) Type() int8 {
	return int8(p) & 7
}
