package internal

// Move
// for later
// [ 8bits chunks]
// [flag][piece][endIdx][startIdx]
type Move struct {
	piece    Piece
	startIdx int8
	endIdx   int8
	flag     int8
}

//type Move struct {
//	piece    Piece
//	startIdx int8
//	endIdx   int8
//}

const (
	NormalMove      = 0
	QueenPromotion  = 1
	KnightPromotion = 2
	BishopPromotion = 3
	RookPromotion   = 4
	EnPassant       = 5
	PawnDoubleMove  = 6
	Castle          = 7
	Capture         = 8
)

func NewMove(piece Piece, startIdx int8, endIdx int8, flag int8) Move {
	return Move{piece, startIdx, endIdx, flag}
}

func (m Move) StartIdx() int8 {
	return m.startIdx
}

func (m Move) EndIdx() int8 {
	return m.endIdx
}

func (m Move) UCI() string {
	startRank, startFile := RankAndFile(m.startIdx)
	endRank, endFile := RankAndFile(m.endIdx)
	var buf [5]byte
	buf[0] = byte('a' + startFile)
	buf[1] = byte('1' + startRank)
	buf[2] = byte('a' + endFile)
	buf[3] = byte('1' + endRank)
	n := 4
	switch m.flag {
	case QueenPromotion:
		buf[4] = 'q'
		n = 5
	case KnightPromotion:
		buf[4] = 'n'
		n = 5
	case BishopPromotion:
		buf[4] = 'b'
		n = 5
	case RookPromotion:
		buf[4] = 'r'
		n = 5
	}
	return string(buf[:n])
}

func isPromotionSquare(color int8, idx int8) bool {
	if color == White {
		return idx >= A8 && idx <= H8
	}

	return idx >= A1 && idx <= H1
}
