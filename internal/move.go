package internal

var promotionFlags = [4]int8{QueenPromotion, KnightPromotion, BishopPromotion, RookPromotion}

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

func isCastleMove(move Move) bool {
	return move.flag == Castle || (move.piece.Type() == King && absInt8(move.EndIdx()-move.StartIdx()) == 2)
}

func isEnPassantMove(pos *Position, move Move) bool {
	if move.flag == EnPassant {
		return true
	}
	if move.piece.Type() != Pawn || pos.enPassantIdx == NoEnPassant || move.EndIdx() != pos.enPassantIdx {
		return false
	}
	return pos.PieceAt(move.EndIdx()) == NoPiece && absInt8(FileFromIdx(move.EndIdx())-FileFromIdx(move.StartIdx())) == 1
}

func classifyMove(pos *Position, piece Piece, startIdx, targetIdx int8) int8 {
	if piece.Type() == King && absInt8(targetIdx-startIdx) == 2 {
		return Castle
	}
	if piece.Type() == Pawn {
		if pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx {
			return EnPassant
		}
		if absInt8(targetIdx-startIdx) == 16 {
			return PawnDoubleMove
		}
	}
	if pos.board[targetIdx] != NoPiece {
		return Capture
	}
	return NormalMove
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
