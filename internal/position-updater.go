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
	p.setPieceAt(oldIdx, NoPiece)
	p.setPieceAt(newIdx, piece)
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
	startPiece := pos.PieceAt(move.StartIdx())
	startPieceType := startPiece.Type()

	history := NewMoveHistory(pos)

	// update position
	updatePieceOnBoard(pos, startPiece, startPieceIdx, endPieceIdx)

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if pos.activeColor == White {
			pos.whiteKingIdx = endPieceIdx
			pos.whiteCastleRights = NoCastle

			if startPieceIdx == E1 {
				if endPieceIdx == G1 {
					updatePieceOnBoard(pos, Piece(White|Rook), H1, F1)
				} else if endPieceIdx == C1 {
					updatePieceOnBoard(pos, Piece(White|Rook), A1, D1)
				}
			}
		} else {
			pos.blackKingIdx = endPieceIdx
			pos.blackCastleRights = NoCastle

			if startPieceIdx == E8 {
				if endPieceIdx == G8 {
					updatePieceOnBoard(pos, Piece(Black|Rook), H8, F8)
				} else if endPieceIdx == C8 {
					updatePieceOnBoard(pos, Piece(Black|Rook), A8, D8)
				}
			}
		}
	}

	// en passant
	if startPieceType == Pawn {
		if endPieceIdx == pos.enPassantIdx {
			var capturesPawnIdx int8
			if pos.activeColor == White {
				capturesPawnIdx = pos.enPassantIdx - 8
			} else {
				capturesPawnIdx = pos.enPassantIdx + 8
			}
			updatePieceOnBoard(pos, Piece(Pawn|pos.activeColor), startPieceIdx, capturesPawnIdx)
			updatePieceOnBoard(pos, Piece(Pawn|pos.activeColor), capturesPawnIdx, endPieceIdx)
		}

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

func (updater *PositionUpdater) UnMakeMove(pos *Position, move Move, history MoveHistory) {
	// Restore the pieces to their original positions
	pos.board = history.board

	// Restore the state of the kings
	pos.whiteKingIdx = history.whiteKingIdx
	pos.blackKingIdx = history.blackKingIdx

	// Restore castling rights
	pos.whiteCastleRights = history.whiteCastleRights
	pos.blackCastleRights = history.blackCastleRights

	// reset king safety
	pos.blackKingSafety = history.blackKingSafety
	pos.whiteKingSafety = history.whiteKingSafety

	// reset occupancies masks
	pos.occupied = history.occupied
	pos.blackOccupied = history.blackOccupied
	pos.whiteOccupied = history.whiteOccupied

	pos.kingBoard = history.kingBoard
	pos.queenBoard = history.queenBoard
	pos.rookBoard = history.rookBoard
	pos.bishopBoard = history.bishopBoard
	pos.knightBoard = history.knightBoard
	pos.pawnBoard = history.pawnBoard

	// Restore en passant index
	pos.enPassantIdx = history.enPassantIdx
	// Restore the active color
	pos.activeColor = history.activeColor
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
