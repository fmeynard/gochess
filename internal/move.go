package internal

import "fmt"

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
	startRank, startFile := RankAndFile(m.StartIdx())
	endRank, endFile := RankAndFile(m.EndIdx())

	uci := fmt.Sprintf(
		"%c%d%c%d",
		'a'+startFile,
		startRank+1,
		'a'+endFile,
		endRank+1,
	)

	switch m.flag {
	case QueenPromotion:
		return uci + "q"
	case KnightPromotion:
		return uci + "n"
	case BishopPromotion:
		return uci + "b"
	case RookPromotion:
		return uci + "r"
	default:
		return uci
	}
}

func isPromotionSquare(color int8, idx int8) bool {
	if color == White {
		return idx >= A8 && idx <= H8
	}

	return idx >= A1 && idx <= H1
}
