package internal

type MoveHistory struct {
	move              Move
	movedPiece        Piece
	capturedPiece     Piece
	captureIdx        int8
	whiteKingIdx      int8
	blackKingIdx      int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantIdx      int8
	activeColor       int8
	whiteKingSafety   int8
	blackKingSafety   int8
	zobristKey        uint64
}

func NewMoveHistory(pos *Position, move Move, movedPiece, capturedPiece Piece, captureIdx int8) MoveHistory {
	return MoveHistory{
		move:              move,
		movedPiece:        movedPiece,
		capturedPiece:     capturedPiece,
		captureIdx:        captureIdx,
		whiteKingIdx:      pos.whiteKingIdx,
		blackKingIdx:      pos.blackKingIdx,
		whiteCastleRights: pos.whiteCastleRights,
		blackCastleRights: pos.blackCastleRights,
		enPassantIdx:      pos.enPassantIdx,
		activeColor:       pos.activeColor,
		blackKingSafety:   pos.blackKingSafety,
		whiteKingSafety:   pos.whiteKingSafety,
		zobristKey:        pos.zobristKey,
	}
}
