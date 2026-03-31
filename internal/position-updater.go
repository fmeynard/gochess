package internal

type PositionUpdater struct {
	moveGenerator *BitsBoardMoveGenerator
}

func kingAffectMask(kingIdx int8) uint64 {
	return queenAttacksMask[kingIdx] | knightAttacksMask[kingIdx] | (uint64(1) << kingIdx)
}

func NewPositionUpdater(moveGenerator *BitsBoardMoveGenerator) *PositionUpdater {
	return &PositionUpdater{
		moveGenerator: moveGenerator,
	}
}

func isCastleMoveForHistory(move Move, piece Piece) bool {
	return piece.Type() == King && absInt8(move.EndIdx()-move.StartIdx()) == 2
}

func isEnPassantMove(pos *Position, move Move, piece Piece) bool {
	if piece.Type() != Pawn || pos.enPassantIdx == NoEnPassant || move.EndIdx() != pos.enPassantIdx {
		return false
	}

	return pos.PieceAt(move.EndIdx()) == NoPiece && absInt8(FileFromIdx(move.EndIdx())-FileFromIdx(move.StartIdx())) == 1
}

func castleRookSquares(activeColor int8, kingEndIdx int8) (int8, int8) {
	if activeColor == White {
		if kingEndIdx == G1 {
			return H1, F1
		}
		return A1, D1
	}

	if kingEndIdx == G8 {
		return H8, F8
	}

	return A8, D8
}

func (updater *PositionUpdater) updateMovesAfterMove(pos *Position, move Move) {
	pos.whiteKingSafety = NotCalculated
	pos.blackKingSafety = NotCalculated
}

func (updater *PositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	startPiece := pos.PieceAt(startPieceIdx)
	startPieceType := startPiece.Type()
	isEnPassant := isEnPassantMove(pos, move, startPiece)

	captureIdx := endPieceIdx
	capturedPiece := pos.PieceAt(endPieceIdx)
	if isEnPassant {
		if pos.activeColor == White {
			captureIdx = endPieceIdx - 8
		} else {
			captureIdx = endPieceIdx + 8
		}
		capturedPiece = pos.PieceAt(captureIdx)
	}

	history := NewMoveHistory(pos, move, startPiece, capturedPiece, captureIdx)

	if isEnPassant {
		pos.removePieceAt(startPieceIdx, startPiece)
		pos.removePieceAt(captureIdx, capturedPiece)
		pos.addPieceAt(endPieceIdx, startPiece)
	} else if capturedPiece != NoPiece {
		pos.capturePiece(startPiece, capturedPiece, startPieceIdx, endPieceIdx)
	} else {
		pos.movePiece(startPiece, startPieceIdx, endPieceIdx)
	}

	switch move.flag {
	case QueenPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Queen))
	case KnightPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Knight))
	case BishopPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Bishop))
	case RookPromotion:
		pos.removePieceAt(endPieceIdx, startPiece)
		pos.addPieceAt(endPieceIdx, Piece(pos.activeColor|Rook))
	}

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if pos.activeColor == White {
			pos.whiteKingIdx = endPieceIdx
			pos.whiteKingAffectMask = kingAffectMask(endPieceIdx)
			pos.whiteCastleRights = NoCastle

			if isCastleMoveForHistory(move, startPiece) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				pos.movePiece(Piece(White|Rook), rookStartIdx, rookEndIdx)
			}
		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackKingAffectMask = kingAffectMask(endPieceIdx)
			pos.blackCastleRights = NoCastle

			if isCastleMoveForHistory(move, startPiece) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				pos.movePiece(Piece(Black|Rook), rookStartIdx, rookEndIdx)
			}
		}
	}

	// en passant
	if startPieceType == Pawn {
		diffIdx := endPieceIdx - startPieceIdx
		if pos.activeColor == White && diffIdx == 16 {
			pos.enPassantIdx = startPieceIdx + 8
		} else if pos.activeColor == Black && diffIdx == -16 {
			pos.enPassantIdx = startPieceIdx - 8
		} else {
			pos.enPassantIdx = NoEnPassant
		}
	} else {
		pos.enPassantIdx = NoEnPassant
	}

	// knight
	if startPieceType == Rook {
		if pos.activeColor == White {
			if startPieceIdx == A1 {
				pos.whiteCastleRights = pos.whiteCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H1 {
				pos.whiteCastleRights = pos.whiteCastleRights &^ KingSideCastle
			}
		} else {
			if startPieceIdx == A8 {
				pos.blackCastleRights = pos.blackCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H8 {
				pos.blackCastleRights = pos.blackCastleRights &^ KingSideCastle
			}
		}
	}

	updater.updateMovesAfterMove(pos, move)

	// change side ( important to do it last for previous updates )
	if pos.activeColor == White {
		pos.activeColor = Black
	} else {
		pos.activeColor = White
	}

	key := history.zobristKey
	key ^= zobristPieceKey(startPiece, startPieceIdx)
	if capturedPiece != NoPiece {
		key ^= zobristPieceKey(capturedPiece, captureIdx)
	}

	finalPiece := startPiece
	switch move.flag {
	case QueenPromotion:
		finalPiece = Piece(history.activeColor | Queen)
	case KnightPromotion:
		finalPiece = Piece(history.activeColor | Knight)
	case BishopPromotion:
		finalPiece = Piece(history.activeColor | Bishop)
	case RookPromotion:
		finalPiece = Piece(history.activeColor | Rook)
	}
	key ^= zobristPieceKey(finalPiece, endPieceIdx)

	if isCastleMoveForHistory(move, startPiece) {
		rookStartIdx, rookEndIdx := castleRookSquares(history.activeColor, endPieceIdx)
		rook := Piece(history.activeColor | Rook)
		key ^= zobristPieceKey(rook, rookStartIdx)
		key ^= zobristPieceKey(rook, rookEndIdx)
	}

	key ^= zobristCastleKey(history.whiteCastleRights, history.blackCastleRights)
	key ^= zobristCastleKey(pos.whiteCastleRights, pos.blackCastleRights)
	key ^= zobristEPKey(history.enPassantIdx)
	key ^= zobristEPKey(pos.enPassantIdx)
	key ^= zobristSideToMove
	pos.zobristKey = key

	return history
}

func (updater *PositionUpdater) UnMakeMove(pos *Position, history MoveHistory) {
	move := history.move
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()

	pos.activeColor = history.activeColor
	pos.whiteKingIdx = history.whiteKingIdx
	pos.blackKingIdx = history.blackKingIdx
	pos.whiteKingAffectMask = kingAffectMask(pos.whiteKingIdx)
	pos.blackKingAffectMask = kingAffectMask(pos.blackKingIdx)
	pos.whiteCastleRights = history.whiteCastleRights
	pos.blackCastleRights = history.blackCastleRights
	pos.whiteKingSafety = history.whiteKingSafety
	pos.blackKingSafety = history.blackKingSafety
	pos.enPassantIdx = history.enPassantIdx

	if move.flag == QueenPromotion || move.flag == KnightPromotion || move.flag == BishopPromotion || move.flag == RookPromotion {
		pos.removePieceAt(endPieceIdx, pos.PieceAt(endPieceIdx))
		pos.addPieceAt(startPieceIdx, history.movedPiece)
	} else {
		pos.movePiece(history.movedPiece, endPieceIdx, startPieceIdx)
	}

	if isCastleMoveForHistory(move, history.movedPiece) {
		rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
		pos.movePiece(Piece(pos.activeColor|Rook), rookEndIdx, rookStartIdx)
	}

	if history.capturedPiece != NoPiece {
		pos.addPieceAt(history.captureIdx, history.capturedPiece)
	}

	pos.zobristKey = history.zobristKey
}

// IsMoveAffectsKing
// The goal here is to trigger king safety recalculation as less as possible,
// but it's a balance, if detection is less performant than the recalculation it's better to recalculate
func (updater *PositionUpdater) IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool {
	var kingAffectMask uint64
	if kingColor == White {
		kingAffectMask = pos.whiteKingAffectMask
	} else {
		kingAffectMask = pos.blackKingAffectMask
	}

	movePiecesMask := uint64(1<<m.startIdx | 1<<m.endIdx)
	return (kingAffectMask & movePiecesMask) != 0
}
