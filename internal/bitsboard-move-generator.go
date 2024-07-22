package internal

const (
	West      int8 = -1
	East      int8 = 1
	South     int8 = -8
	North     int8 = 8
	SouthWest int8 = -9
	SouthEast int8 = -7
	NorthWest int8 = 7
	NorthEast int8 = 9
)

var (
	QueenDirections  = []int8{West, East, South, North, SouthWest, SouthEast, NorthWest, NorthEast}
	RookDirections   = []int8{South, North, West, East}
	BishopDirections = []int8{SouthWest, SouthEast, NorthWest, NorthEast}

	// move offsets per piece type

	KingOffsets   = []int8{-9, -8, -7, -1, 1, 7, 8, 9}
	KnightOffsets = []int8{-17, -15, -10, -6, 6, 10, 15, 17}

	// move per rank and file

	kingMoves = [8][2]int8{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	knightMoves = [8][2]int8{
		{2, 1}, {1, 2}, {-1, 2}, {-2, 1},
		{-2, -1}, {-1, -2}, {1, -2}, {2, -1},
	}

	sliderAttackMasks           [64][8]uint64
	queenAttacksMask            [64]uint64
	knightAttacksMask           [64]uint64
	kingAttacksMask             [64]uint64
	diagonalAttacksMask         [64][4]uint64
	diagonalCombinedAttacksMask [64]uint64
)

type BitsBoardMoveGenerator struct {
	bishopMasks            [64][4][]int8
	rookMasks              [64][4][]int8
	knightMasks            [64]uint64
	kingMasks              [64]uint64
	whitePawnMovesMasks    [64]uint64
	blackPawnMovesMasks    [64]uint64
	whitePawnCapturesMasks [64]uint64
	blackPawnCapturesMasks [64]uint64
	sliderAttackMasks      [64][8]uint64
}

func NewBitsBoardMoveGenerator() *BitsBoardMoveGenerator {
	bitsBoardMoveGenerator := &BitsBoardMoveGenerator{}
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
		g.initSliderAttackMasks()
		g.initKnightAttacksMasks()
		g.initKingAttacksMasks()
		g.initDiagonalAttacksMasks()
	}
}

func (g *BitsBoardMoveGenerator) initDiagonalAttacksMasks() {
	for idx := int8(0); idx < 64; idx++ {
		for dirIdx, dir := range BishopDirections {
			mask := uint64(0)
			for step := int8(1); step < 8; step++ {
				nextSquare := idx + step*dir
				if nextSquare < 0 || nextSquare >= 64 || !isSameLineOrRow(idx, nextSquare, dir) {
					break
				}
				mask |= 1 << nextSquare
			}
			diagonalAttacksMask[idx][dirIdx] = mask
			diagonalCombinedAttacksMask[idx] |= mask
		}
	}
}

func (g *BitsBoardMoveGenerator) initSliderAttackMasks() {
	for idx := int8(0); idx < 64; idx++ {
		for dirIdx, dir := range QueenDirections {
			mask := uint64(0)
			for step := int8(1); step < 8; step++ {
				nextSquare := idx + step*dir
				if nextSquare < 0 || nextSquare >= 64 || !isSameLineOrRow(idx, nextSquare, dir) {
					break
				}
				mask |= 1 << nextSquare
			}
			g.sliderAttackMasks[idx][dirIdx] = mask
			queenAttacksMask[idx] |= mask
		}
	}

	sliderAttackMasks = g.sliderAttackMasks
}

func (g *BitsBoardMoveGenerator) initKnightAttacksMasks() {
	for idx := int8(0); idx < 64; idx++ {
		rank, file := RankAndFile(idx)
		for _, move := range knightMoves {
			newFile := file + move[0]
			newRank := rank + move[1]
			if isOnBoard(newFile, newRank) {
				knightAttacksMask[idx] |= uint64(1 << (newRank*8 + newFile))
			}
		}
	}
}

func (g *BitsBoardMoveGenerator) initKingAttacksMasks() {
	for idx := int8(0); idx < 64; idx++ {
		rank, file := RankAndFile(idx)

		for _, move := range kingMoves {
			newFile := file + move[0]
			newRank := rank + move[1]
			if isOnBoard(newFile, newRank) {
				kingAttacksMask[idx] |= uint64(1 << (newRank*8 + newFile))
			}
		}
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
	// SouthWest direction
	for r, f := squareRank+1, squareFile-1; r < 8 && f >= 0; r, f = r+1, f-1 {
		g.bishopMasks[squareIdx][0] = append(g.bishopMasks[squareIdx][0], r*8+f)
	}
	// SouthEast direction
	for r, f := squareRank+1, squareFile+1; r < 8 && f < 8; r, f = r+1, f+1 {
		g.bishopMasks[squareIdx][1] = append(g.bishopMasks[squareIdx][1], r*8+f)
	}
	// NorthWest direction
	for r, f := squareRank-1, squareFile-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		g.bishopMasks[squareIdx][2] = append(g.bishopMasks[squareIdx][2], r*8+f)
	}
	// NorthEast direction
	for r, f := squareRank-1, squareFile+1; r >= 0 && f < 8; r, f = r-1, f+1 {
		g.bishopMasks[squareIdx][3] = append(g.bishopMasks[squareIdx][3], r*8+f)
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
	for _, offset := range KnightOffsets {
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
	for _, offset := range KingOffsets {
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
func (g *BitsBoardMoveGenerator) PawnPseudoLegalMoves(pos *Position, idx int8) ([]int8, int8) {
	moves := make([]int8, 0, 4)
	promotionIdx := int8(-1)

	isWhite := pos.activeColor == White

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

		if pos.occupied&(1<<targetIdx) == 0 {
			if (isWhite && targetIdx >= A8) || (!isWhite && targetIdx <= H1) {
				// Handle promotion
				promotionIdx = targetIdx
			} else {
				if isWhite && targetIdx >= A4 && targetIdx <= H4 && targetIdx-idx == 16 && pos.occupied&(1<<(idx+8)) != 0 {
					continue
				}

				if !isWhite && targetIdx >= A5 && targetIdx <= H5 && idx-targetIdx == 16 && pos.occupied&(1<<(idx-8)) != 0 {
					continue
				}

				moves = append(moves, targetIdx)
			}
		}
	}

	// Handle en passant
	// check first as next part is modifying the captureMask
	// no color check need as pos.enPassantIdx is automatically set to NoEnPassant after player move,
	// so it is impossible to be in the scenario of a Pawn being able
	if pos.enPassantIdx != NoEnPassant {
		if captureMask&(1<<pos.enPassantIdx) != 0 {
			moves = append(moves, pos.enPassantIdx)
		}
	}

	// Generate captures
	oMask := pos.OpponentOccupiedMask()
	for captureMask != 0 {
		targetIdx := leastSignificantOne(captureMask)
		captureMask &= captureMask - 1

		if (oMask & (1 << targetIdx)) != 0 {
			if (isWhite && targetIdx >= A8) || (!isWhite && targetIdx <= H1) {
				// Handle promotion with capture
				promotionIdx = targetIdx
			} else {
				moves = append(moves, targetIdx)
			}
		}
	}

	return moves, promotionIdx
}

func (g *BitsBoardMoveGenerator) KingPseudoLegalMoves(pos *Position, idx int8) []int8 {
	var moves = make([]int8, 0, 8)

	opponentOccupiedMask := pos.OpponentOccupiedMask()
	castleRights := pos.CastleRights()
	isWhite := pos.activeColor == White
	moveMask := g.kingMasks[idx]
	for moveMask != 0 {
		targetIdx := leastSignificantOne(moveMask)
		moveMask &= moveMask - 1

		if pos.occupied&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			moves = append(moves, targetIdx)
		}
	}

	var (
		kingStartIdx int8
	)
	if isWhite {
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
		if isWhite {
			queenCastleIdx = C1
			queenPathIsClear = (pos.occupied&(1<<B1) == 0) && (pos.occupied&(1<<C1) == 0) && (pos.occupied&(1<<D1) == 0)
		} else {
			queenCastleIdx = C8
			queenPathIsClear = (pos.occupied&(1<<B8) == 0) && (pos.occupied&(1<<C8) == 0) && (pos.occupied&(1<<D8) == 0)
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
		if isWhite {
			kingCastleIdx = G1
			kingPathIsClear = (pos.occupied&(1<<G1) == 0) && (pos.occupied&(1<<F1) == 0)
		} else {
			kingCastleIdx = G8
			kingPathIsClear = (pos.occupied&(1<<G8) == 0) && (pos.occupied&(1<<F8) == 0)
		}

		if kingPathIsClear {
			moves = append(moves, kingCastleIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) KnightPseudoLegalMoves(pos *Position, idx int8) []int8 {
	var moves = make([]int8, 0, 8)
	knightMask := g.knightMasks[idx]
	opponentOccupiedMask := pos.OpponentOccupiedMask()

	for knightMask != 0 {
		targetIdx := leastSignificantOne(knightMask)
		knightMask &= knightMask - 1

		if pos.occupied&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			moves = append(moves, targetIdx)
		}
	}

	return moves
}

func (g *BitsBoardMoveGenerator) SliderPseudoLegalMoves(pos *Position, idx int8, pieceType int8) []int8 {
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

	opponentOccupiedMask := pos.OpponentOccupiedMask()
	for dir := 0; dir < 4; dir++ {
		if processRookDirections {
			for _, targetIdx := range g.rookMasks[idx][dir] {
				targetMask := uint64(1 << targetIdx)
				if pos.occupied&targetMask == 0 {
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
				if pos.occupied&targetMask == 0 {
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
