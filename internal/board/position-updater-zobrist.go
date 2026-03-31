package board

type ZobristPositionUpdater struct {
	inner *PlainPositionUpdater
}

func NewZobristPositionUpdater(inner *PlainPositionUpdater) *ZobristPositionUpdater {
	return &ZobristPositionUpdater{inner: inner}
}

func (updater *ZobristPositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	history := updater.inner.MakeMove(pos, move)
	history.zobristKey = pos.zobristKey

	startPiece := move.piece
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	capturedPiece := history.capturedPiece
	captureIdx := history.captureIdx

	key := history.zobristKey
	key ^= zobristPieceKey(startPiece, startPieceIdx)
	if capturedPiece != NoPiece {
		key ^= zobristPieceKey(capturedPiece, captureIdx)
	}

	finalPiece := startPiece
	switch move.flag {
	case QueenPromotion:
		finalPiece = Piece(startPiece.Color() | Queen)
	case KnightPromotion:
		finalPiece = Piece(startPiece.Color() | Knight)
	case BishopPromotion:
		finalPiece = Piece(startPiece.Color() | Bishop)
	case RookPromotion:
		finalPiece = Piece(startPiece.Color() | Rook)
	}
	key ^= zobristPieceKey(finalPiece, endPieceIdx)

	if isCastleMove(move) {
		rookStartIdx, rookEndIdx := castleRookSquares(startPiece.Color(), endPieceIdx)
		rook := Piece(startPiece.Color() | Rook)
		key ^= zobristPieceKey(rook, rookStartIdx)
		key ^= zobristPieceKey(rook, rookEndIdx)
	}

	key ^= zobristCastleKey(history.whiteCastleRights(), history.blackCastleRights())
	key ^= zobristCastleKey(pos.whiteCastleRights, pos.blackCastleRights)
	key ^= zobristEPKey(history.enPassantIdx())
	key ^= zobristEPKey(pos.enPassantIdx)
	key ^= zobristSideToMove
	pos.zobristKey = key

	return history
}

func (updater *ZobristPositionUpdater) UnMakeMove(pos *Position, history MoveHistory) {
	updater.inner.UnMakeMove(pos, history)
	pos.zobristKey = history.zobristKey
}
