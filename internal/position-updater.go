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
	//p.board[oldIdx] = NoPiece
	//p.board[newIdx] = piece
	//
	//if piece.Color() == White {
	//	p.whiteOccupied &= ^(uint64(1) << oldIdx)
	//	p.whiteOccupied |= uint64(1) << newIdx
	//	p.blackOccupied &= ^(uint64(1) << newIdx)
	//} else {
	//	p.whiteOccupied &= ^(uint64(1) << newIdx)
	//	p.blackOccupied &= ^(uint64(1) << oldIdx)
	//	p.blackOccupied |= uint64(1) << newIdx
	//}
	//
	//p.occupied &= ^(uint64(1) << oldIdx)
	//p.occupied |= uint64(1) << newIdx
	p.setPieceAt(oldIdx, NoPiece)
	p.setPieceAt(newIdx, piece)
}

func (updater *PositionUpdater) CopyPosition(initPos *Position) *Position {
	newPos := *initPos

	return &newPos
}

func (updater *PositionUpdater) MakeMove(pos *Position, move Move) MoveHistory {
	startPieceIdx := move.StartIdx()
	endPieceIdx := move.EndIdx()
	startPiece := pos.PieceAt(move.StartIdx())
	endPiece := pos.PieceAt(move.EndIdx())
	startPieceType := startPiece.Type()

	history := MoveHistory{
		startPiece:        startPiece,
		startIdx:          move.StartIdx(),
		endIdx:            move.EndIdx(),
		capturedPiece:     endPiece,
		whiteKingIdx:      pos.whiteKingIdx,
		blackKingIdx:      pos.blackKingIdx,
		whiteCastleRights: pos.whiteCastleRights,
		blackCastleRights: pos.blackCastleRights,
		enPassantIdx:      pos.enPassantIdx,
		activeColor:       pos.activeColor,
	}

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
	pos.setPieceAt(history.startIdx, history.startPiece)
	pos.setPieceAt(history.endIdx, history.capturedPiece)

	// Restore the state of the kings
	pos.whiteKingIdx = history.whiteKingIdx
	pos.blackKingIdx = history.blackKingIdx

	// Restore castling rights
	pos.whiteCastleRights = history.whiteCastleRights
	pos.blackCastleRights = history.blackCastleRights

	if move.piece.Type() == King {
		if move.startIdx == E1 && move.piece.Color() == White {
			if move.endIdx == C1 {
				pos.setPieceAt(D1, NoPiece)
				pos.setPieceAt(A1, Piece(Rook|White))
			}
			if move.endIdx == G1 {
				pos.setPieceAt(F1, NoPiece)
				pos.setPieceAt(H1, Piece(Rook|White))
			}
		}

		if move.startIdx == E8 && move.piece.Color() == Black {
			if move.endIdx == C8 {
				pos.setPieceAt(D8, NoPiece)
				pos.setPieceAt(A8, Piece(Rook|Black))
			}
			if move.endIdx == G8 {
				pos.setPieceAt(F8, NoPiece)
				pos.setPieceAt(H8, Piece(Rook|Black))
			}
		}
	}

	// Restore en passant index
	pos.enPassantIdx = history.enPassantIdx
	if history.enPassantIdx != NoEnPassant && move.endIdx == history.enPassantIdx {
		if pos.activeColor == White {
			pos.setPieceAt(history.endIdx-8, Piece(Black|Pawn))
		} else {
			pos.setPieceAt(history.endIdx+8, Piece(White|Pawn))
		}
	}

	// Restore the active color
	pos.activeColor = history.activeColor
}

type MoveHistory struct {
	capturedPiece     Piece
	startPiece        Piece
	startIdx          int8
	endIdx            int8
	whiteKingIdx      int8
	blackKingIdx      int8
	whiteCastleRights int8
	blackCastleRights int8
	enPassantIdx      int8
	activeColor       int8
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
