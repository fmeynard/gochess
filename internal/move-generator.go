package internal

// direction offsets
const (
	LEFT      int8 = -1
	RIGHT     int8 = 1
	UP        int8 = -8
	DOWN      int8 = 8
	UpLeft    int8 = -9
	UpRight   int8 = -7
	DownLeft  int8 = 7
	DownRight int8 = 9
)

var (
	QueenDirections  = []int8{LEFT, RIGHT, UP, DOWN, UpLeft, UpRight, DownLeft, DownRight}
	RookDirections   = []int8{UP, DOWN, LEFT, RIGHT}
	BishopDirections = []int8{UpLeft, UpRight, DownLeft, DownRight}
)

var kingMoves2 = []int8{
	-9, -8, -7, -1, 1, 7, 8, 9,
}

var knightMoves2 = []int8{
	-17, -15, -10, -6,
	6, 10, 15, 17,
}

var (
	queenDirections = [8]int8{8, -8, 1, -1, 9, -9, 7, -7}
	kingMoves       = [8][2]int8{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
)

var knightMoves = [8][2]int8{
	{2, 1}, {1, 2}, {-1, 2}, {-2, 1},
	{-2, -1}, {-1, -2}, {1, -2}, {2, -1},
}

func (p *Position) IsCheck() bool {
	return IsKingInCheck(p, p.activeColor)
}
