package movegen

import (
	board "chessV2/internal/board"
	"math/bits"
)

func appendNormalMovesFromMask(piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx
		dst[count] = board.NewMove(piece, startIdx, targetIdx, board.NormalMove)
		count++
	}

	return count
}

func appendCaptureMovesFromMask(piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx
		dst[count] = board.NewMove(piece, startIdx, targetIdx, board.Capture)
		count++
	}

	return count
}

func appendPawnQuietMoves(piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx

		flag := int8(board.NormalMove)
		if absInt8(targetIdx-startIdx) == 16 {
			flag = board.PawnDoubleMove
		}

		dst[count] = board.NewMove(piece, startIdx, targetIdx, flag)
		count++
	}

	return count
}

func appendPromotionMoves(piece Piece, startIdx int8, targets uint64, dst []Move, count int) int {
	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx
		for _, flag := range promotionFlags {
			dst[count] = board.NewMove(piece, startIdx, targetIdx, flag)
			count++
		}
	}
	return count
}

func sliderTargetsMask(pos *Position, idx int8, pieceType int8, friendlyOcc uint64) uint64 {
	switch pieceType {
	case Bishop:
		return bishopAttacksMagic(idx, pos.Occupied()) &^ friendlyOcc
	case Rook:
		return rookAttacksMagic(idx, pos.Occupied()) &^ friendlyOcc
	default:
		return (rookAttacksMagic(idx, pos.Occupied()) | bishopAttacksMagic(idx, pos.Occupied())) &^ friendlyOcc
	}
}

const (
	notAFile uint64 = 0xFEFEFEFEFEFEFEFE
	notHFile uint64 = 0x7F7F7F7F7F7F7F7F
)

func (g *PseudoLegalMoveGenerator) appendKingMoves(pos *Position, piece Piece, kingIdx int8, enemyColor int8, enemyKingMask uint64, info *positionAnalysis, dst []Move, count int) int {
	friendlyOcc := pos.OccupancyMask(pos.ActiveColor())
	targets := kingAttacksMask[kingIdx] &^ friendlyOcc &^ enemyKingMask
	occWithoutKing := pos.Occupied() &^ (uint64(1) << kingIdx)

	var enemyOcc uint64
	if enemyColor == White {
		enemyOcc = pos.WhiteOccupied()
	} else {
		enemyOcc = pos.BlackOccupied()
	}

	enemyPawns := pos.PawnBoard() & enemyOcc
	var pawnAttacks uint64
	if enemyColor == White {
		pawnAttacks = ((enemyPawns & notAFile) << 7) | ((enemyPawns & notHFile) << 9)
	} else {
		pawnAttacks = ((enemyPawns & notHFile) >> 7) | ((enemyPawns & notAFile) >> 9)
	}

	var knightAttacks uint64
	enemyKnights := pos.KnightBoard() & enemyOcc
	for enemyKnights != 0 {
		sq := int8(bits.TrailingZeros64(enemyKnights))
		enemyKnights &^= 1 << sq
		knightAttacks |= knightAttacksMask[sq]
	}

	var enemyKingIdx int8
	if enemyColor == White {
		enemyKingIdx = pos.WhiteKingIdx()
	} else {
		enemyKingIdx = pos.BlackKingIdx()
	}

	targets &^= pawnAttacks | knightAttacks | kingAttacksMask[enemyKingIdx]

	rookQueens := (pos.RookBoard() | pos.QueenBoard()) & enemyOcc
	bishopQueens := (pos.BishopBoard() | pos.QueenBoard()) & enemyOcc

	for targets != 0 {
		targetIdx := int8(bits.TrailingZeros64(targets))
		targets &^= 1 << targetIdx

		occ := occWithoutKing &^ (uint64(1) << targetIdx)
		if rookQueens != 0 && rookAttacksMagic(targetIdx, occ)&rookQueens != 0 {
			continue
		}
		if bishopQueens != 0 && bishopAttacksMagic(targetIdx, occ)&bishopQueens != 0 {
			continue
		}

		flag := int8(board.NormalMove)
		if pos.PieceAt(targetIdx) != NoPiece {
			flag = board.Capture
		}
		dst[count] = board.NewMove(piece, kingIdx, targetIdx, flag)
		count++
	}

	if info.inCheck {
		return count
	}

	castleRights := pos.CastleRights()
	if pos.ActiveColor() == White && kingIdx == E1 {
		if castleRights&KingSideCastle != 0 &&
			pos.PieceAt(F1) == NoPiece && pos.PieceAt(G1) == NoPiece &&
			!isSquareAttacked(pos, F1, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, G1, enemyColor, occWithoutKing&^(uint64(1)<<G1)) {
			dst[count] = board.NewMove(piece, E1, G1, board.Castle)
			count++
		}
		if castleRights&QueenSideCastle != 0 &&
			pos.PieceAt(B1) == NoPiece && pos.PieceAt(C1) == NoPiece && pos.PieceAt(D1) == NoPiece &&
			!isSquareAttacked(pos, D1, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, C1, enemyColor, occWithoutKing&^(uint64(1)<<C1)) {
			dst[count] = board.NewMove(piece, E1, C1, board.Castle)
			count++
		}
		return count
	}

	if pos.ActiveColor() == Black && kingIdx == E8 {
		if castleRights&KingSideCastle != 0 &&
			pos.PieceAt(F8) == NoPiece && pos.PieceAt(G8) == NoPiece &&
			!isSquareAttacked(pos, F8, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, G8, enemyColor, occWithoutKing&^(uint64(1)<<G8)) {
			dst[count] = board.NewMove(piece, E8, G8, board.Castle)
			count++
		}
		if castleRights&QueenSideCastle != 0 &&
			pos.PieceAt(B8) == NoPiece && pos.PieceAt(C8) == NoPiece && pos.PieceAt(D8) == NoPiece &&
			!isSquareAttacked(pos, D8, enemyColor, occWithoutKing) &&
			!isSquareAttacked(pos, C8, enemyColor, occWithoutKing&^(uint64(1)<<C8)) {
			dst[count] = board.NewMove(piece, E8, C8, board.Castle)
			count++
		}
	}

	return count
}

func (g *PseudoLegalMoveGenerator) appendPawnMoves(pos *Position, positionUpdater board.MoveApplier, piece Piece, idx int8, enemyOcc uint64, info *positionAnalysis, pinRay uint64, isPinned bool, inCheckColor int8, dst []Move, count int) int {
	var quietTargets uint64
	var captureTargets uint64
	var promotionTargets uint64
	var epTarget uint64

	occ := pos.Occupied()
	if pos.ActiveColor() == White {
		oneStep := idx + 8
		if occ&(uint64(1)<<oneStep) == 0 {
			quietTargets |= uint64(1) << oneStep
			if idx >= A2 && idx <= H2 && occ&(uint64(1)<<(idx+16)) == 0 {
				quietTargets |= uint64(1) << (idx + 16)
			}
		}
		captureTargets = g.whitePawnCapturesMasks[idx] & enemyOcc
		if pos.EnPassantIdx() != NoEnPassant && g.whitePawnCapturesMasks[idx]&(uint64(1)<<pos.EnPassantIdx()) != 0 {
			epTarget = uint64(1) << pos.EnPassantIdx()
		}
	} else {
		oneStep := idx - 8
		if occ&(uint64(1)<<oneStep) == 0 {
			quietTargets |= uint64(1) << oneStep
			if idx >= A7 && idx <= H7 && occ&(uint64(1)<<(idx-16)) == 0 {
				quietTargets |= uint64(1) << (idx - 16)
			}
		}
		captureTargets = g.blackPawnCapturesMasks[idx] & enemyOcc
		if pos.EnPassantIdx() != NoEnPassant && g.blackPawnCapturesMasks[idx]&(uint64(1)<<pos.EnPassantIdx()) != 0 {
			epTarget = uint64(1) << pos.EnPassantIdx()
		}
	}

	if info != nil && info.checkerCount == 1 {
		evasionMask := info.evasionMask
		quietTargets &= evasionMask
		captureTargets &= evasionMask
	}
	if isPinned {
		quietTargets &= pinRay
		captureTargets &= pinRay
	}

	allTargets := quietTargets | captureTargets
	if info != nil && info.checkerCount == 1 {
		allTargets &= info.evasionMask
	}

	if pos.ActiveColor() == White {
		promotionTargets = allTargets & (uint64(0xFF) << 56)
	} else {
		promotionTargets = allTargets & uint64(0xFF)
	}
	quietTargets &^= promotionTargets
	captureTargets &^= promotionTargets

	count = appendPawnQuietMoves(piece, idx, quietTargets, dst, count)
	count = appendCaptureMovesFromMask(piece, idx, captureTargets, dst, count)
	count = appendPromotionMoves(piece, idx, promotionTargets, dst, count)

	if epTarget == 0 {
		return count
	}
	if isPinned && pinRay&epTarget == 0 {
		return count
	}

	move := board.NewMove(piece, idx, int8(bits.TrailingZeros64(epTarget)), board.EnPassant)
	history := positionUpdater.MakeMove(pos, move)
	if !IsKingInCheck(pos, inCheckColor) {
		dst[count] = move
		count++
	}
	positionUpdater.UnMakeMove(pos, history)
	return count
}

func (g *PseudoLegalMoveGenerator) appendNonKingMovesNoCheck(pos *Position, positionUpdater board.MoveApplier, piecesMask uint64, friendlyOcc uint64, enemyOccNoKing uint64, enemyKingMask uint64, info *positionAnalysis, dst []Move, count int) int {
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))
		piecesMask &^= 1 << idx

		piece := pos.PieceAt(idx)
		pieceType := piece.Type()
		pinRay := info.pinRayBySq[idx]

		switch pieceType {
		case Pawn:
			count = g.appendPawnMoves(pos, positionUpdater, piece, idx, enemyOccNoKing, info, pinRay, info.pinnedMask&(1<<idx) != 0, pos.ActiveColor(), dst, count)
		case Knight:
			if pinRay != 0 {
				continue
			}
			targets := knightAttacksMask[idx] &^ friendlyOcc &^ enemyKingMask
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		case Bishop, Rook, Queen:
			targets := sliderTargetsMask(pos, idx, pieceType, friendlyOcc) &^ enemyKingMask
			if pinRay != 0 {
				targets &= pinRay
			}
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		}
	}
	return count
}

func (g *PseudoLegalMoveGenerator) appendNonKingMovesNoCheckNoPins(pos *Position, positionUpdater board.MoveApplier, piecesMask uint64, friendlyOcc uint64, enemyOccNoKing uint64, enemyKingMask uint64, dst []Move, count int) int {
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))
		piecesMask &^= 1 << idx

		piece := pos.PieceAt(idx)
		switch piece.Type() {
		case Pawn:
			count = g.appendPawnMoves(pos, positionUpdater, piece, idx, enemyOccNoKing, nil, 0, false, pos.ActiveColor(), dst, count)
		case Knight:
			targets := knightAttacksMask[idx] &^ friendlyOcc &^ enemyKingMask
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		case Bishop, Rook, Queen:
			targets := sliderTargetsMask(pos, idx, piece.Type(), friendlyOcc) &^ enemyKingMask
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		}
	}
	return count
}

func (g *PseudoLegalMoveGenerator) appendNonKingMovesInCheck(pos *Position, positionUpdater board.MoveApplier, piecesMask uint64, friendlyOcc uint64, enemyOccNoKing uint64, enemyKingMask uint64, info *positionAnalysis, dst []Move, count int) int {
	evasionMask := info.evasionMask
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))
		piecesMask &^= 1 << idx

		piece := pos.PieceAt(idx)
		pieceType := piece.Type()
		pinRay := info.pinRayBySq[idx]

		switch pieceType {
		case Pawn:
			count = g.appendPawnMoves(pos, positionUpdater, piece, idx, enemyOccNoKing, info, pinRay, info.pinnedMask&(1<<idx) != 0, pos.ActiveColor(), dst, count)
		case Knight:
			if pinRay != 0 {
				continue
			}
			targets := knightAttacksMask[idx] &^ friendlyOcc &^ enemyKingMask & evasionMask
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		case Bishop, Rook, Queen:
			targets := sliderTargetsMask(pos, idx, pieceType, friendlyOcc) &^ enemyKingMask & evasionMask
			if pinRay != 0 {
				targets &= pinRay
			}
			count = appendNormalMovesFromMask(piece, idx, targets&^enemyOccNoKing, dst, count)
			count = appendCaptureMovesFromMask(piece, idx, targets&enemyOccNoKing, dst, count)
		}
	}
	return count
}

func (g *PseudoLegalMoveGenerator) LegalMovesInto(pos *Position, positionUpdater board.MoveApplier, dst []Move) int {
	count := 0
	color := pos.ActiveColor()
	kingIdx := pos.BlackKingIdx()
	friendlyOcc := pos.BlackOccupied()
	enemyOcc := pos.WhiteOccupied()
	enemyColor := White
	enemyKingIdx := pos.WhiteKingIdx()
	if color == White {
		kingIdx = pos.WhiteKingIdx()
		friendlyOcc = pos.WhiteOccupied()
		enemyOcc = pos.BlackOccupied()
		enemyColor = Black
		enemyKingIdx = pos.BlackKingIdx()
	}
	enemyKingMask := uint64(1) << enemyKingIdx
	enemyOccNoKing := enemyOcc &^ enemyKingMask

	kingPiece := pos.PieceAt(kingIdx)
	if kingPiece != Piece(color|King) {
		return 0
	}

	var info positionAnalysis
	computePositionAnalysis(pos, kingIdx, friendlyOcc, enemyOcc, &info)
	count = g.appendKingMoves(pos, kingPiece, kingIdx, enemyColor, enemyKingMask, &info, dst, count)

	if info.checkerCount >= 2 {
		return count
	}

	piecesMask := friendlyOcc &^ (uint64(1) << kingIdx)
	if info.checkerCount == 0 {
		if info.pinnedMask == 0 {
			return g.appendNonKingMovesNoCheckNoPins(pos, positionUpdater, piecesMask, friendlyOcc, enemyOccNoKing, enemyKingMask, dst, count)
		}
		return g.appendNonKingMovesNoCheck(pos, positionUpdater, piecesMask, friendlyOcc, enemyOccNoKing, enemyKingMask, &info, dst, count)
	}

	return g.appendNonKingMovesInCheck(pos, positionUpdater, piecesMask, friendlyOcc, enemyOccNoKing, enemyKingMask, &info, dst, count)
}
