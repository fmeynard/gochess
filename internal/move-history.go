package internal

type MoveHistory struct {
	whiteKingIdx      int8
	blackKingIdx      int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantIdx      int8
	activeColor       int8
	whiteKingSafety   int8
	blackKingSafety   int8
	whiteOccupied     uint64
	blackOccupied     uint64
	occupied          uint64

	queenBoard  uint64
	kingBoard   uint64
	bishopBoard uint64
	rookBoard   uint64
	knightBoard uint64
	pawnBoard   uint64
}

func NewMoveHistory(pos *Position) MoveHistory {
	return MoveHistory{
		whiteKingIdx:      pos.whiteKingIdx,
		blackKingIdx:      pos.blackKingIdx,
		whiteCastleRights: pos.whiteCastleRights,
		blackCastleRights: pos.blackCastleRights,
		enPassantIdx:      pos.enPassantIdx,
		activeColor:       pos.activeColor,
		blackKingSafety:   pos.blackKingSafety,
		whiteKingSafety:   pos.whiteKingSafety,
		whiteOccupied:     pos.whiteOccupied,
		blackOccupied:     pos.blackOccupied,
		occupied:          pos.occupied,
		kingBoard:         pos.kingBoard,
		queenBoard:        pos.queenBoard,
		rookBoard:         pos.rookBoard,
		bishopBoard:       pos.bishopBoard,
		knightBoard:       pos.knightBoard,
		pawnBoard:         pos.pawnBoard,
	}
}
