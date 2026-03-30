package internal

type PositionUpdater struct {
	moveGenerator IMoveGenerator
}

func NewPositionUpdater(moveGenerator IMoveGenerator) *PositionUpdater {
	return &PositionUpdater{
		moveGenerator: moveGenerator,
	}
}

// updatePieceOnBoard
// update board and position masks
// Important: capture moves need to update opponent occupancy maks
func updatePieceOnBoard(p *Position, piece Piece, oldIdx int8, newIdx int8) {
	capturedPiece := p.PieceAt(newIdx)
	p.removePieceAt(oldIdx, piece)
	p.removePieceAt(newIdx, capturedPiece)
	p.addPieceAt(newIdx, piece)
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
	if updater.IsMoveAffectsKing(pos, move, White) {

		pos.whiteKingSafety = NotCalculated
		if IsKingInCheck(pos, White) {
			pos.whiteKingSafety = KingIsCheck
		} else {
			pos.whiteKingSafety = KingIsSafe
		}
	}

	if updater.IsMoveAffectsKing(pos, move, Black) {
		pos.blackKingSafety = NotCalculated
		if IsKingInCheck(pos, Black) {
			pos.blackKingSafety = KingIsCheck
		} else {
			pos.blackKingSafety = KingIsSafe
		}
	}
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

	// update position
	if isEnPassant {
		pos.removePieceAt(startPieceIdx, startPiece)
		pos.removePieceAt(captureIdx, capturedPiece)
		pos.addPieceAt(endPieceIdx, startPiece)
	} else {
		updatePieceOnBoard(pos, startPiece, startPieceIdx, endPieceIdx)
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
			pos.whiteCastleRights = NoCastle

			if isCastleMoveForHistory(move, startPiece) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				updatePieceOnBoard(pos, Piece(White|Rook), rookStartIdx, rookEndIdx)
			}
		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackCastleRights = NoCastle

			if isCastleMoveForHistory(move, startPiece) {
				rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
				updatePieceOnBoard(pos, Piece(Black|Rook), rookStartIdx, rookEndIdx)
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

	return history
}

func (updater *PositionUpdater) UnMakeMove(pos *Position, history MoveHistory) {
	move := history.move
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()

	pos.activeColor = history.activeColor
	pos.whiteKingIdx = history.whiteKingIdx
	pos.blackKingIdx = history.blackKingIdx
	pos.whiteCastleRights = history.whiteCastleRights
	pos.blackCastleRights = history.blackCastleRights
	pos.whiteKingSafety = history.whiteKingSafety
	pos.blackKingSafety = history.blackKingSafety
	pos.enPassantIdx = history.enPassantIdx

	if move.flag == QueenPromotion || move.flag == KnightPromotion || move.flag == BishopPromotion || move.flag == RookPromotion {
		pos.removePieceAt(endPieceIdx, pos.PieceAt(endPieceIdx))
		pos.addPieceAt(startPieceIdx, history.movedPiece)
	} else {
		updatePieceOnBoard(pos, history.movedPiece, endPieceIdx, startPieceIdx)
	}

	if isCastleMoveForHistory(move, history.movedPiece) {
		rookStartIdx, rookEndIdx := castleRookSquares(pos.activeColor, endPieceIdx)
		updatePieceOnBoard(pos, Piece(pos.activeColor|Rook), rookEndIdx, rookStartIdx)
	}

	if history.capturedPiece != NoPiece {
		pos.addPieceAt(history.captureIdx, history.capturedPiece)
	}
}

// IsMoveAffectsKing
// The goal here is to trigger king safety recalculation as less as possible,
// but it's a balance, if detection is less performant than the recalculation it's better to recalculate
func (updater *PositionUpdater) IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool {
	// Convert indices to row and column
	var kingIdx int8
	if kingColor == White {
		kingIdx = pos.whiteKingIdx
	} else {
		kingIdx = pos.blackKingIdx
	}

	movePiecesMask := uint64(1<<m.startIdx | 1<<m.endIdx)
	// Direct involvement
	if m.StartIdx() == kingIdx || m.EndIdx() == kingIdx {
		return true
	}

	if (queenAttacksMask[kingIdx] & movePiecesMask) != 0 {
		return true
	}

	// Special checks for knights moves
	if knightAttacksMask[kingIdx]&movePiecesMask != 0 {
		return true
	}

	return false
}
