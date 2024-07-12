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

func (m Move) Flag() int8 {
	return m.flag
}

func (m Move) UCI() string {
	startRank, startFile := RankAndFile(m.StartIdx())
	endRank, endFile := RankAndFile(m.EndIdx())

	return fmt.Sprintf(
		"%c%d%c%d",
		'a'+startFile,
		startRank+1,
		'a'+endFile,
		endRank+1,
	)
}
