package internal

import "sync"

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
	attackTablesOnce sync.Once
	QueenDirections  = []int8{West, East, South, North, SouthWest, SouthEast, NorthWest, NorthEast}
	RookDirections   = []int8{South, North, West, East}
	BishopDirections = []int8{SouthWest, SouthEast, NorthWest, NorthEast}

	// move offsets per piece type

	KingOffsets   = []int8{-9, -8, -7, -1, 1, 7, 8, 9}
	KnightOffsets = []int8{-17, -15, -10, -6, 6, 10, 15, 17}

	rayDirections               = [8]int8{West, East, South, North, SouthWest, SouthEast, NorthWest, NorthEast}
	sliderAttackMasks           [64][8]uint64
	queenAttacksMask            [64]uint64
	orthogonalAttacksMask       [64]uint64
	knightAttacksMask           [64]uint64
	kingAttacksMask             [64]uint64
	diagonalAttacksMask         [64][4]uint64
	diagonalCombinedAttacksMask [64]uint64
	betweenMasks                [64][64]uint64
)

type PseudoLegalMoveGenerator struct {
	bishopMasks            [64][4][7]int8
	bishopMaskLens         [64][4]int8
	rookMasks              [64][4][7]int8
	rookMaskLens           [64][4]int8
	knightMasks            [64]uint64
	kingMasks              [64]uint64
	whitePawnMovesMasks    [64]uint64
	blackPawnMovesMasks    [64]uint64
	whitePawnCapturesMasks [64]uint64
	blackPawnCapturesMasks [64]uint64
	sliderAttackMasks      [64][8]uint64
}

func init() {
	ensureAttackTables()
}

func NewPseudoLegalMoveGenerator() *PseudoLegalMoveGenerator {
	ensureAttackTables()
	bitsBoardMoveGenerator := &PseudoLegalMoveGenerator{}
	bitsBoardMoveGenerator.initPieceMasks()

	return bitsBoardMoveGenerator
}

func ensureAttackTables() {
	attackTablesOnce.Do(func() {
		g := &PseudoLegalMoveGenerator{}
		g.initPieceMasks()
		g.initSliderAttackMasks()
		g.initKnightAttacksMasks()
		g.initKingAttacksMasks()
		g.initDiagonalAttacksMasks()
		initMagicBitboards()
	})
}

func (g *PseudoLegalMoveGenerator) initPieceMasks() {
	for squareIdx := int8(0); squareIdx < 64; squareIdx++ {
		g.initBishopMaskForSquare(squareIdx)
		g.initRookMaskForSquare(squareIdx)
		g.initKnightMaskForSquare(squareIdx)
		g.initKingMaskForSquare(squareIdx)
		g.initPawnMasksForSquare(squareIdx)
	}
}

func (g *PseudoLegalMoveGenerator) initDiagonalAttacksMasks() {
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

func (g *PseudoLegalMoveGenerator) initSliderAttackMasks() {
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
			if dirIdx < 4 {
				orthogonalAttacksMask[idx] |= mask
			}
		}
	}

	sliderAttackMasks = g.sliderAttackMasks
	initBetweenMasks()
}

func initBetweenMasks() {
	for from := int8(0); from < 64; from++ {
		for to := int8(0); to < 64; to++ {
			if from == to {
				continue
			}
			mask := uint64(0)
			for dirIdx, dir := range rayDirections {
				ray := sliderAttackMasks[from][dirIdx]
				if ray&(uint64(1)<<to) == 0 {
					continue
				}
				next := from + dir
				for next != to {
					mask |= uint64(1) << next
					next += dir
				}
				mask &^= uint64(1) << to
				betweenMasks[from][to] = mask
				break
			}
		}
	}
}

func (g *PseudoLegalMoveGenerator) initKnightAttacksMasks() {

	knightMoves := [8][2]int8{
		{2, 1}, {1, 2}, {-1, 2}, {-2, 1},
		{-2, -1}, {-1, -2}, {1, -2}, {2, -1},
	}

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

func (g *PseudoLegalMoveGenerator) initKingAttacksMasks() {
	kingMoves := [8][2]int8{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

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
func (g *PseudoLegalMoveGenerator) initPawnMasksForSquare(squareIdx int8) {
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

func (g *PseudoLegalMoveGenerator) initBishopMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	n := int8(0)
	for r, f := squareRank+1, squareFile-1; r < 8 && f >= 0; r, f = r+1, f-1 {
		g.bishopMasks[squareIdx][0][n] = r*8 + f
		n++
	}
	g.bishopMaskLens[squareIdx][0] = n

	n = 0
	for r, f := squareRank+1, squareFile+1; r < 8 && f < 8; r, f = r+1, f+1 {
		g.bishopMasks[squareIdx][1][n] = r*8 + f
		n++
	}
	g.bishopMaskLens[squareIdx][1] = n

	n = 0
	for r, f := squareRank-1, squareFile-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		g.bishopMasks[squareIdx][2][n] = r*8 + f
		n++
	}
	g.bishopMaskLens[squareIdx][2] = n

	n = 0
	for r, f := squareRank-1, squareFile+1; r >= 0 && f < 8; r, f = r-1, f+1 {
		g.bishopMasks[squareIdx][3][n] = r*8 + f
		n++
	}
	g.bishopMaskLens[squareIdx][3] = n
}

func (g *PseudoLegalMoveGenerator) initRookMaskForSquare(squareIdx int8) {
	squareRank, squareFile := RankAndFile(squareIdx)
	n := int8(0)
	for r := squareRank + 1; r < 8; r++ {
		g.rookMasks[squareIdx][0][n] = r*8 + squareFile
		n++
	}
	g.rookMaskLens[squareIdx][0] = n

	n = 0
	for r := squareRank - 1; r >= 0; r-- {
		g.rookMasks[squareIdx][1][n] = r*8 + squareFile
		n++
	}
	g.rookMaskLens[squareIdx][1] = n

	n = 0
	for f := squareFile - 1; f >= 0; f-- {
		g.rookMasks[squareIdx][2][n] = squareRank*8 + f
		n++
	}
	g.rookMaskLens[squareIdx][2] = n

	n = 0
	for f := squareFile + 1; f < 8; f++ {
		g.rookMasks[squareIdx][3][n] = squareRank*8 + f
		n++
	}
	g.rookMaskLens[squareIdx][3] = n
}

func (g *PseudoLegalMoveGenerator) initKnightMaskForSquare(squareIdx int8) {
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

func (g *PseudoLegalMoveGenerator) initKingMaskForSquare(squareIdx int8) {
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
func (g *PseudoLegalMoveGenerator) PawnPseudoLegalMoves(pos *Position, idx int8) ([]int8, int8) {
	var buf [4]int8
	count, promotionIdx := g.PawnPseudoLegalMovesInto(pos, idx, buf[:])
	moves := make([]int8, count)
	copy(moves, buf[:count])
	return moves, promotionIdx
}

func (g *PseudoLegalMoveGenerator) PawnPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) (int, int8) {
	count := 0
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

	for moveMask != 0 {
		targetIdx := leastSignificantOne(moveMask)
		moveMask &= moveMask - 1

		if pos.occupied&(1<<targetIdx) == 0 {
			if (isWhite && targetIdx >= A8) || (!isWhite && targetIdx <= H1) {
				promotionIdx = targetIdx
			}

			if isWhite && targetIdx >= A4 && targetIdx <= H4 && targetIdx-idx == 16 && pos.occupied&(1<<(idx+8)) != 0 {
				continue
			}

			if !isWhite && targetIdx >= A5 && targetIdx <= H5 && idx-targetIdx == 16 && pos.occupied&(1<<(idx-8)) != 0 {
				continue
			}

			dst[count] = targetIdx
			count++
		}
	}

	if pos.enPassantIdx != NoEnPassant && captureMask&(1<<pos.enPassantIdx) != 0 {
		dst[count] = pos.enPassantIdx
		count++
	}

	oMask := pos.OpponentOccupiedMask()
	for captureMask != 0 {
		targetIdx := leastSignificantOne(captureMask)
		captureMask &= captureMask - 1

		if (oMask & (1 << targetIdx)) != 0 {
			if (isWhite && targetIdx >= A8) || (!isWhite && targetIdx <= H1) {
				promotionIdx = targetIdx
			}

			dst[count] = targetIdx
			count++
		}
	}

	return count, promotionIdx
}

func (g *PseudoLegalMoveGenerator) KingPseudoLegalMoves(pos *Position, idx int8) []int8 {
	var buf [8]int8
	count := g.KingPseudoLegalMovesInto(pos, idx, buf[:])
	moves := make([]int8, count)
	copy(moves, buf[:count])
	return moves
}

func (g *PseudoLegalMoveGenerator) KingPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) int {
	count := 0
	opponentOccupiedMask := pos.OpponentOccupiedMask()
	castleRights := pos.CastleRights()
	isWhite := pos.activeColor == White
	moveMask := g.kingMasks[idx]
	for moveMask != 0 {
		targetIdx := leastSignificantOne(moveMask)
		moveMask &= moveMask - 1

		if pos.occupied&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			dst[count] = targetIdx
			count++
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
		return count
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
			dst[count] = queenCastleIdx
			count++
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
			dst[count] = kingCastleIdx
			count++
		}
	}

	return count
}

func (g *PseudoLegalMoveGenerator) KnightPseudoLegalMoves(pos *Position, idx int8) []int8 {
	var buf [8]int8
	count := g.KnightPseudoLegalMovesInto(pos, idx, buf[:])
	moves := make([]int8, count)
	copy(moves, buf[:count])
	return moves
}

func (g *PseudoLegalMoveGenerator) KnightPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) int {
	count := 0
	knightMask := g.knightMasks[idx]
	opponentOccupiedMask := pos.OpponentOccupiedMask()

	for knightMask != 0 {
		targetIdx := leastSignificantOne(knightMask)
		knightMask &= knightMask - 1

		if pos.occupied&(1<<targetIdx) == 0 || opponentOccupiedMask&(1<<targetIdx) != 0 {
			dst[count] = targetIdx
			count++
		}
	}

	return count
}

func (g *PseudoLegalMoveGenerator) SliderPseudoLegalMoves(pos *Position, idx int8, pieceType int8) []int8 {
	var buf [28]int8
	count := g.SliderPseudoLegalMovesInto(pos, idx, pieceType, buf[:])
	moves := make([]int8, count)
	copy(moves, buf[:count])
	return moves
}

func (g *PseudoLegalMoveGenerator) SliderPseudoLegalMovesInto(pos *Position, idx int8, pieceType int8, dst []int8) int {
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

	count := 0

	opponentOccupiedMask := pos.OpponentOccupiedMask()
	for dir := 0; dir < 4; dir++ {
		if processRookDirections {
			rayLen := int(g.rookMaskLens[idx][dir])
			for j := 0; j < rayLen; j++ {
				targetIdx := g.rookMasks[idx][dir][j]
				targetMask := uint64(1 << targetIdx)
				if pos.occupied&targetMask == 0 {
					dst[count] = targetIdx
					count++
					continue
				}

				if opponentOccupiedMask&targetMask != 0 {
					dst[count] = targetIdx
					count++
				}

				break
			}
		}

		if processBishopDirections {
			rayLen := int(g.bishopMaskLens[idx][dir])
			for j := 0; j < rayLen; j++ {
				targetIdx := g.bishopMasks[idx][dir][j]
				targetMask := uint64(1 << targetIdx)
				if pos.occupied&targetMask == 0 {
					dst[count] = targetIdx
					count++
					continue
				}

				if opponentOccupiedMask&targetMask != 0 {
					dst[count] = targetIdx
					count++
				}

				break
			}
		}
	}

	_ = maxMovesCnt
	return count
}
