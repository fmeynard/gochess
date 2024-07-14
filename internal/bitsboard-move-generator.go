package internal

type BitsBoardMoveGenerator struct {
	bishopMasks            [64]uint64
	rookMasks              [64]uint64
	knightMasks            [64]uint64
	kingMasks              [64]uint64
	whitePawnMovesMasks    [64]uint64
	blackPawnMovesMasks    [64]uint64
	whitePawnCapturesMasks [64]uint64
	blackPawnCapturesMasks [64]uint64

	rookDirections   []int8
	bishopDirections []int8
	knightOffsets    []int8
	kingOffsets      []int8
}

func NewBitsBoardMoveGenerator() *BitsBoardMoveGenerator {
	bitsBoardMoveGenerator := &BitsBoardMoveGenerator{
		knightOffsets:    []int8{-17, -15, -10, -6, 6, 10, 15, 17},
		kingOffsets:      []int8{-9, -8, -7, -1, 1, 7, 8, 9},
		bishopDirections: []int8{UpLeft, UpRight, DownLeft, DownRight},
		rookDirections:   []int8{UP, DOWN, LEFT, RIGHT},
	}
	bitsBoardMoveGenerator.initMasks()

	return bitsBoardMoveGenerator
}

// Initialize the attack masks
func (g *BitsBoardMoveGenerator) initMasks() {
	for squareIdx := int8(0); squareIdx < 64; squareIdx++ {
		g.initBishopMaskForSquare(squareIdx)
		g.initRookMaskForSquare(squareIdx)
		g.initKnightMaskForSquare(squareIdx)
		g.initKingMaskForSquare(squareIdx)
		g.initPawnMasksForSquare(squareIdx)
	}
}

// initPawnMasksForSquare
// Pawns only moves forward and their capture patterns are different from their move patterns
// En Passant needs to be checked separately as its tight to the position
func (g *BitsBoardMoveGenerator) initPawnMasksForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)

	// White pawn single move
	if squareRank < 7 {
		g.whitePawnMovesMasks[squareIdx] |= 1 << (squareIdx + 8)
	}
	// White pawn double move
	if squareRank == 1 {
		g.whitePawnMovesMasks[squareIdx] |= 1 << (squareIdx + 16)
	}
	// White pawn captures
	if squareRank < 7 && squareFile > 0 {
		g.whitePawnCapturesMasks[squareIdx] |= 1 << (squareIdx + 7)
	}
	if squareRank < 7 && squareFile < 7 {
		g.whitePawnCapturesMasks[squareIdx] |= 1 << (squareIdx + 9)
	}

	// Black pawn single move
	if squareRank > 0 {
		g.blackPawnMovesMasks[squareIdx] |= 1 << (squareIdx - 8)
	}
	// Black pawn double move
	if squareRank == 6 {
		g.blackPawnMovesMasks[squareIdx] |= 1 << (squareIdx - 16)
	}
	// Black pawn captures
	if squareRank > 0 && squareFile > 0 {
		g.blackPawnCapturesMasks[squareIdx] |= 1 << (squareIdx - 9)
	}
	if squareRank > 0 && squareFile < 7 {
		g.blackPawnCapturesMasks[squareIdx] |= 1 << (squareIdx - 7)
	}
}

func (g *BitsBoardMoveGenerator) initBishopMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	for _, dir := range g.bishopDirections {
		mask := uint64(0)
		prevRank, prevFile := squareRank, squareFile
		for targetIdx := squareIdx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
			targetRank, targetFile := RankAndFile(targetIdx)

			// only possible with out-of-bounds/cross-boards move
			if targetRank == squareRank ||
				targetFile == squareFile ||
				absInt8(prevRank-targetRank) != 1 ||
				absInt8(prevFile-targetFile) != 1 {
				break
			}

			mask |= 1 << targetIdx

			prevRank, prevFile = targetRank, targetFile
		}
		g.bishopMasks[squareIdx] |= mask
	}
}

func (g *BitsBoardMoveGenerator) initRookMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	for _, dir := range g.rookDirections {
		mask := uint64(0)
		for targetIdx := squareIdx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
			targetRank, targetFile := RankAndFile(targetIdx)
			if targetFile != squareFile && targetRank != squareRank {
				break
			}
			mask |= 1 << targetIdx

			if (targetFile == 0 && dir == LEFT) ||
				(targetFile == 7 && dir == RIGHT) ||
				(targetRank == 0 && dir == UP) ||
				(targetRank == 7 && dir == DOWN) {
				break
			}
		}
		g.rookMasks[squareIdx] |= mask
	}
}

func (g *BitsBoardMoveGenerator) initKnightMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	for _, offset := range g.knightOffsets {
		targetIdx := squareIdx + offset
		if targetIdx >= 0 && targetIdx < 64 {
			targetRank, targetFile := RankAndFile(targetIdx)
			fileDiff := absInt8(squareFile - targetFile)
			rankDiff := absInt8(squareRank - targetRank)
			if fileDiff <= 2 && rankDiff <= 2 {
				g.knightMasks[squareIdx] |= 1 << targetIdx
			}
		}
	}
}

func (g *BitsBoardMoveGenerator) initKingMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	for _, offset := range g.kingOffsets {
		targetIdx := squareIdx + offset
		if targetIdx >= 0 && targetIdx < 64 {
			targetRank, targetFile := RankAndFile(targetIdx)
			fileDiff := absInt8(squareFile - targetFile)
			rankDiff := absInt8(squareRank - targetRank)
			if fileDiff <= 1 && rankDiff <= 1 {
				g.kingMasks[squareIdx] |= 1 << targetIdx
			}
		}
	}
}

// PawnPseudoLegalMoves Generate pawn moves using bitboards
func (g *BitsBoardMoveGenerator) PawnPseudoLegalMoves(pos Position, idx int8) ([]int8, int8) {
	moves := make([]int8, 0, 4)
	promotionIdx := int8(-1)

	pieceColor := pos.board[idx].Color()
	isWhite := pieceColor == White

	var moveMask, captureMask uint64
	if isWhite {
		moveMask = g.whitePawnMovesMasks[idx]
		captureMask = g.whitePawnCapturesMasks[idx]
	} else {
		moveMask = g.blackPawnMovesMasks[idx]
		captureMask = g.blackPawnCapturesMasks[idx]
	}

	// Generate moves

	for moveMask != 0 {
		targetIdx := leastSignificantOne(moveMask)
		moveMask &= moveMask - 1

		if pos.board[targetIdx] == NoPiece {
			targetRank := RankFromIdx(targetIdx)
			if (isWhite && targetIdx/8 == 7) || (!isWhite && targetIdx/8 == 0) {
				// Handle promotion
				promotionIdx = targetIdx
			} else {
				if targetRank == 3 && pieceColor == White && targetIdx-idx == 16 && pos.board[idx+8] != NoPiece {
					continue
				}

				if targetRank == 4 && pieceColor == Black && idx-targetIdx == 16 && pos.board[idx-8] != NoPiece {
					continue
				}
				moves = append(moves, targetIdx)
			}
		}
	}

	// Handle en passant
	// check first as next part is modifying the captureMask
	if pos.enPassantIdx != NoEnPassant {
		enPassantTarget := pos.enPassantIdx
		if captureMask&(1<<enPassantTarget) != 0 {
			moves = append(moves, enPassantTarget)
		}
	}

	// Generate captures
	for captureMask != 0 {
		targetIdx := leastSignificantOne(captureMask)
		captureMask &= captureMask - 1
		target := pos.board[targetIdx]

		if target != NoPiece && pieceColor != target.Color() {
			targetRank := RankFromIdx(targetIdx)
			if (isWhite && targetRank == 7) || (!isWhite && targetRank == 0) {
				// Handle promotion with capture
				promotionIdx = targetIdx
			} else {
				moves = append(moves, targetIdx)
			}
		}
	}

	return moves, promotionIdx
}

func (g *BitsBoardMoveGenerator) KingPseudoLegalMoves(pos Position, idx int8) []int8 {
	var moves = make([]int8, 0, 8)

	pieceColor := pos.board[idx].Color()
	for _, currentOffset := range g.kingOffsets {
		targetIdx := idx + currentOffset
		if targetIdx < 0 || targetIdx > 63 || g.kingMasks[idx]&(1<<targetIdx) == 0 {
			continue
		}

		target := pos.board[targetIdx]
		if target == NoPiece || target.Color() != pieceColor {
			moves = append(moves, targetIdx)
		}
	}

	var (
		castleRights int8
		kingStartIdx int8
	)
	if pos.activeColor == White {
		castleRights = pos.whiteCastleRights
		kingStartIdx = E1
	} else {
		castleRights = pos.blackCastleRights
		kingStartIdx = E8
	}

	// early exit no castle
	if idx != kingStartIdx || castleRights == NoCastle {
		return moves
	}

	// queen side
	var (
		queenPathIsClear bool
		queenRookIdx     int8
		queenCastleIdx   int8
	)
	if (castleRights & QueenSideCastle) != 0 {
		if pos.activeColor == White {
			queenRookIdx = A1
			queenCastleIdx = C1
			queenPathIsClear = (pos.PieceAt(B1) == NoPiece) && (pos.PieceAt(C1) == NoPiece) && (pos.PieceAt(D1) == NoPiece)
		} else {
			queenRookIdx = A8
			queenCastleIdx = C8
			queenPathIsClear = (pos.PieceAt(B8) == NoPiece) && (pos.PieceAt(C8) == NoPiece) && (pos.PieceAt(D8) == NoPiece)
		}

		if queenPathIsClear && pos.PieceAt(queenRookIdx).Type() == Rook {
			moves = append(moves, queenCastleIdx)
		}
	}

	// king side
	var (
		kingPathIsClear bool
		kingRookIdx     int8
		kingCastleIdx   int8
	)
	if (castleRights & KingSideCastle) != 0 {
		if pos.activeColor == White {
			kingRookIdx = H1
			kingCastleIdx = G1
			kingPathIsClear = (pos.PieceAt(G1) == NoPiece) && (pos.PieceAt(F1) == NoPiece)
		} else {
			kingRookIdx = H8
			kingCastleIdx = G8
			kingPathIsClear = (pos.PieceAt(G8) == NoPiece) && (pos.PieceAt(F8) == NoPiece)
		}

		if kingPathIsClear && pos.PieceAt(kingRookIdx).Type() == Rook {
			moves = append(moves, kingCastleIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) KnightPseudoLegalMoves(pos Position, idx int8) []int8 {
	var moves = make([]int8, 0, 8)

	pieceColor := pos.board[idx].Color()
	for _, currentOffset := range g.knightOffsets {
		targetIdx := idx + currentOffset
		if targetIdx < 0 || targetIdx > 63 || g.knightMasks[idx]&(1<<targetIdx) == 0 {
			continue
		}

		target := pos.board[targetIdx]
		if target == NoPiece || target.Color() != pieceColor {
			moves = append(moves, targetIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) SliderPseudoLegalMoves(pos Position, idx int8) []int8 {
	var (
		moves                   = make([]int8, 0, 28)
		processBishopDirections = false
		processRookDirections   = false
	)

	piece := pos.board[idx]

	switch piece.Type() {
	case Bishop:
		processBishopDirections = true
	case Rook:
		processRookDirections = true
	case Queen:
		processBishopDirections = true
		processRookDirections = true
	}

	pieceRank, pieceFile := RankAndFile(idx)
	pieceColor := piece.Color()

	if processBishopDirections {
		for _, dir := range BishopDirections {
			prevRank, prevFile := pieceRank, pieceFile
			for targetIdx := idx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
				moveToTargetMask := g.bishopMasks[idx] & (1 << targetIdx)
				if moveToTargetMask == 0 {
					break
				}

				targetRank, targetFile := RankAndFile(targetIdx)
				if targetRank == pieceRank ||
					targetFile == pieceFile ||
					absInt8(prevRank-targetRank) != 1 ||
					absInt8(prevFile-targetFile) != 1 {
					break
				}

				if (moveToTargetMask & pos.occupied) != 0 {
					if pos.board[targetIdx].Color() != pieceColor {
						moves = append(moves, targetIdx)
					}
					break
				}
				moves = append(moves, targetIdx)

				prevRank, prevFile = targetRank, targetFile
			}
		}
	}

	if processRookDirections {
		for _, dir := range RookDirections {
			for targetIdx := idx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
				moveToTargetMask := g.rookMasks[idx] & (1 << targetIdx)
				if moveToTargetMask == 0 {
					break
				}

				if (moveToTargetMask & pos.occupied) != 0 {
					if pos.board[targetIdx].Color() != pieceColor {
						moves = append(moves, targetIdx)
					}
					break
				}
				moves = append(moves, targetIdx)

				targetFile := FileFromIdx(targetIdx)
				if (targetFile == 0 && dir == LEFT) || (targetFile == 7 && dir == RIGHT) {
					break
				}

				targetRank := RankFromIdx(targetIdx)
				if (targetRank == 0 && dir == UP) || (targetRank == 7 && dir == DOWN) {
					break
				}
			}
		}
	}

	return moves
}
