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
	p.board[oldIdx] = NoPiece
	p.board[newIdx] = piece

	if piece.Color() == White {
		p.whiteOccupied &= ^(uint64(1) << oldIdx)
		p.whiteOccupied |= uint64(1) << newIdx
		p.blackOccupied &= ^(uint64(1) << newIdx)
	} else {
		p.whiteOccupied &= ^(uint64(1) << newIdx)
		p.blackOccupied &= ^(uint64(1) << oldIdx)
		p.blackOccupied |= uint64(1) << newIdx
	}

	p.occupied &= ^(uint64(1) << oldIdx)
	p.occupied |= uint64(1) << newIdx
}

func (updater *PositionUpdater) CopyPosition(initPos *Position) *Position {
	newPos := *initPos

	return &newPos
}

func (updater *PositionUpdater) PositionAfterMove(initPos *Position, move Move) *Position {
	newPos := updater.CopyPosition(initPos)

	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	startPiece := newPos.PieceAt(move.StartIdx())
	startPieceType := startPiece.Type()

	// update position
	updatePieceOnBoard(newPos, startPiece, startPieceIdx, endPieceIdx)

	// King move -> update king pos and castleRights
	if startPieceType == King {
		if newPos.activeColor == White {
			newPos.whiteKingIdx = endPieceIdx
			newPos.whiteCastleRights = NoCastle

			if startPieceIdx == E1 {
				if endPieceIdx == G1 {
					updatePieceOnBoard(newPos, Piece(White|Rook), H1, F1)
				} else if endPieceIdx == C1 {
					updatePieceOnBoard(newPos, Piece(White|Rook), A1, D1)
				}
			}
		} else {
			newPos.blackKingIdx = endPieceIdx
			newPos.blackCastleRights = NoCastle

			if startPieceIdx == E8 {
				if endPieceIdx == G8 {
					updatePieceOnBoard(newPos, Piece(Black|Rook), H8, F8)
				} else if endPieceIdx == C8 {
					updatePieceOnBoard(newPos, Piece(Black|Rook), A8, D8)
				}
			}
		}
	}

	// en passant
	if startPieceType == Pawn {
		if endPieceIdx == newPos.enPassantIdx {
			var capturesPawnIdx int8
			if newPos.activeColor == White {
				capturesPawnIdx = newPos.enPassantIdx - 8
			} else {
				capturesPawnIdx = newPos.enPassantIdx + 8
			}
			updatePieceOnBoard(newPos, Piece(Pawn|newPos.activeColor), startPieceIdx, capturesPawnIdx)
			updatePieceOnBoard(newPos, Piece(Pawn|newPos.activeColor), capturesPawnIdx, endPieceIdx)
		}

		diffIdx := endPieceIdx - startPieceIdx
		if newPos.activeColor == White && diffIdx == 16 {
			newPos.enPassantIdx = startPieceIdx + 8
		} else if newPos.activeColor == Black && diffIdx == -16 {
			newPos.enPassantIdx = startPieceIdx - 8
		} else {
			newPos.enPassantIdx = NoEnPassant
		}
	} else {
		newPos.enPassantIdx = NoEnPassant
	}

	// knight
	if startPieceType == Rook {
		if newPos.activeColor == White {
			if startPieceIdx == A1 {
				newPos.whiteCastleRights = newPos.whiteCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H1 {
				newPos.whiteCastleRights = newPos.whiteCastleRights &^ KingSideCastle
			}
		} else {
			if startPieceIdx == A8 {
				newPos.blackCastleRights = newPos.blackCastleRights &^ QueenSideCastle
			} else if startPieceIdx == H8 {
				newPos.blackCastleRights = newPos.blackCastleRights &^ KingSideCastle
			}
		}
	}

	// change side ( important to do it last for previous updates )
	if newPos.activeColor == White {
		newPos.activeColor = Black
	} else {
		newPos.activeColor = White
	}

	//updateAttackVectors(&newPos)

	return newPos
}

//
//func (updater *PositionUpdater) updateMovesAfterMove(pos *Position, move Move) {
//	// Identify pieces potentially affected by the move
//	fromIdx := move.StartIdx()
//	toIdx := move.EndIdx()
//
//	for _, dir := range []int8{8, -8, 1, -1, 7, -7, 9, -9} {
//		for dist := int8(1); dist <= 7; dist++ { // Maximum board distance
//			affectedIdx := fromIdx + dir*dist
//			if affectedIdx < 0 || affectedIdx >= 64 || !isSameLine(fromIdx, affectedIdx, dir) {
//				break
//			}
//			piece := pos.board[affectedIdx]
//			if piece != NoPiece && canReach(piece.Type(), dir) {
//				//delete(pos.moveCache, affectedIdx)
//				//pos.generateMovesForPiece(affectedIdx)
//				pos.movesCache[affectedIdx] = nil
//				pos.generateMovesForPiece(affectedIdx)
//			}
//		}
//	}
//	pos.moveCache[toIdx] = pos.generateMovesForPiece(toIdx) // Recalculate moves for the moved piece
//}
