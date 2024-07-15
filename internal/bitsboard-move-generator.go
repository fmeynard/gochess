package internal

type BitsBoardMoveGenerator struct {
	bishopMasks            [64][4][]int8
	rookMasks              [64][4][]int8
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
		var d int
		switch dir {
		case UpLeft:
			d = 0
		case UpRight:
			d = 1
		case DownLeft:
			d = 2
		case DownRight:
			d = 3
		}

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

			g.bishopMasks[squareIdx][d] = append(g.bishopMasks[squareIdx][d], targetIdx)

			prevRank, prevFile = targetRank, targetFile
		}
	}
}

func (g *BitsBoardMoveGenerator) initRookMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	// Up direction
	for r := squareRank + 1; r < 8; r++ {
		g.rookMasks[squareIdx][0] = append(g.rookMasks[squareIdx][0], r*8+squareFile)
	}
	// Down direction
	for r := squareRank - 1; r >= 0; r-- {
		g.rookMasks[squareIdx][1] = append(g.rookMasks[squareIdx][1], r*8+squareFile)
	}
	// Left direction
	for f := squareFile - 1; f >= 0; f-- {
		g.rookMasks[squareIdx][2] = append(g.rookMasks[squareIdx][2], squareRank*8+f)
	}
	// Right direction
	for f := squareFile + 1; f < 8; f++ {
		g.rookMasks[squareIdx][3] = append(g.rookMasks[squareIdx][3], squareRank*8+f)
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
func (g *BitsBoardMoveGenerator) PawnPseudoLegalMoves(idx int8, activeColor int8, enPassantIdx int8, occupiedMask uint64, opponentOccupiedMask uint64) ([]int8, int8) {
	moves := make([]int8, 0, 4)
	promotionIdx := int8(-1)

	isWhite := activeColor == White

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

		if occupiedMask&(1<<targetIdx) == 0 {
			targetRank := RankFromIdx(targetIdx)
			if (isWhite && targetRank == 7) || (!isWhite && targetRank == 0) {
				// Handle promotion
				promotionIdx = targetIdx
			} else {
				if targetRank == 3 && activeColor == White && targetIdx-idx == 16 && occupiedMask&(1<<(idx+8)) != 0 {
					continue
				}

				if targetRank == 4 && activeColor == Black && idx-targetIdx == 16 && occupiedMask&(1<<(idx-8)) != 0 {
					continue
				}
				moves = append(moves, targetIdx)
			}
		}
	}

	// Handle en passant
	// check first as next part is modifying the captureMask
	if enPassantIdx != NoEnPassant {
		enPassantTarget := enPassantIdx
		if captureMask&(1<<enPassantTarget) != 0 {
			moves = append(moves, enPassantTarget)
		}
	}

	// Generate captures
	for captureMask != 0 {
		targetIdx := leastSignificantOne(captureMask)
		captureMask &= captureMask - 1

		if opponentOccupiedMask&(1<<targetIdx) != 0 {
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

func (g *BitsBoardMoveGenerator) KingPseudoLegalMoves(
	idx int8, activeColor int8,
	castleRights int8,
	occupiedMask uint64,
	opponentOccupiedMask uint64,
) []int8 {
	var moves = make([]int8, 0, 8)

	moveMask := g.kingMasks[idx]
	for moveMask != 0 {
		targetIdx := leastSignificantOne(moveMask)
		moveMask &= moveMask - 1

		if occupiedMask&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			moves = append(moves, targetIdx)
		}
	}

	var (
		kingStartIdx int8
	)
	if activeColor == White {
		kingStartIdx = E1
	} else {
		kingStartIdx = E8
	}

	// early exit no castle
	if idx != kingStartIdx || castleRights == NoCastle {
		return moves
	}

	// queen side
	var (
		queenPathIsClear bool
		queenCastleIdx   int8
	)
	if (castleRights & QueenSideCastle) != 0 {
		if activeColor == White {
			queenCastleIdx = C1
			queenPathIsClear = (occupiedMask&(1<<B1) == 0) && (occupiedMask&(1<<C1) == 0) && (occupiedMask&(1<<D1) == 0)
		} else {
			queenCastleIdx = C8
			queenPathIsClear = (occupiedMask&(1<<B8) == 0) && (occupiedMask&(1<<C8) == 0) && (occupiedMask&(1<<D8) == 0)
		}

		if queenPathIsClear {
			moves = append(moves, queenCastleIdx)
		}
	}

	// king side
	var (
		kingPathIsClear bool
		kingCastleIdx   int8
	)
	if (castleRights & KingSideCastle) != 0 {
		if activeColor == White {
			kingCastleIdx = G1
			kingPathIsClear = (occupiedMask&(1<<G1) == 0) && (occupiedMask&(1<<F1) == 0)
		} else {
			kingCastleIdx = G8
			kingPathIsClear = (occupiedMask&(1<<G8) == 0) && (occupiedMask&(1<<F8) == 0)
		}

		if kingPathIsClear {
			moves = append(moves, kingCastleIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) KnightPseudoLegalMoves(
	idx int8,
	occupiedMask uint64,
	opponentOccupiedMask uint64,
) []int8 {
	var moves = make([]int8, 0, 8)
	knightMask := g.knightMasks[idx]

	for knightMask != 0 {
		targetIdx := leastSignificantOne(knightMask)
		knightMask &= knightMask - 1

		if occupiedMask&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			moves = append(moves, targetIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) SliderPseudoLegalMoves(
	idx int8,
	pieceType int8,
	occupiedMask uint64,
	opponentOccupiedMask uint64,
) []int8 {
	var (
		processBishopDirections = false
		processRookDirections   = false
	)

	maxMovesCnt := 28
	switch pieceType {
	case Bishop:
		processBishopDirections = true
		maxMovesCnt = 14
	case Rook:
		processRookDirections = true
		maxMovesCnt = 14
	case Queen:
		processBishopDirections = true
		processRookDirections = true
	}

	moves := make([]int8, 0, maxMovesCnt)

	for dir := 0; dir < 4; dir++ {
		if processRookDirections {
			for _, targetIdx := range g.rookMasks[idx][dir] {
				targetMask := uint64(1 << targetIdx)
				if occupiedMask&targetMask == 0 {
					moves = append(moves, targetIdx)
					continue
				}

				if opponentOccupiedMask&targetMask != 0 {
					moves = append(moves, targetIdx)
				}

				break
			}
		}

		if processBishopDirections {
			for _, targetIdx := range g.bishopMasks[idx][dir] {
				targetMask := uint64(1 << targetIdx)
				if occupiedMask&targetMask == 0 {
					moves = append(moves, targetIdx)
					continue
				}

				if opponentOccupiedMask&targetMask != 0 {
					moves = append(moves, targetIdx)
				}

				break
			}
		}
	}

	return moves
}
