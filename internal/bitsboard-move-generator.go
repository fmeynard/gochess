package internal

type BitsBoardMoveGenerator struct {
	bishopMasks [64]uint64
	rookMasks   [64]uint64

	rookDirections   []int8
	bishopDirections []int8
}

func NewBitsBoardMoveGenerator() *BitsBoardMoveGenerator {
	bitsBoardMoveGenerator := &BitsBoardMoveGenerator{}
	bitsBoardMoveGenerator.initMasks()

	return bitsBoardMoveGenerator
}

// Initialize the attack masks
func (g *BitsBoardMoveGenerator) initMasks() {
	for squareIdx := int8(0); squareIdx < 64; squareIdx++ {
		//squareFile := FileFromIdx(squareIdx)
		squareRank, squareFile := RankAndFile(squareIdx)
		for _, dir := range BishopDirections {
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

		for _, dir := range RookDirections {
			mask := uint64(0)
			for targetIdx := squareIdx + dir; targetIdx >= 0 && targetIdx < 64; targetIdx += dir {
				targetRank := RankFromIdx(targetIdx)
				targetFile := FileFromIdx(targetIdx)
				if targetFile != squareFile && targetRank != squareRank {
					break
				}
				mask |= 1 << targetIdx

				if (targetFile == 0 && dir == LEFT) || (targetFile == 7 && dir == RIGHT) {
					break
				}

				if (targetRank == 0 && dir == UP) || (targetRank == 7 && dir == DOWN) {
					break
				}
			}
			g.rookMasks[squareIdx] |= mask
		}
	}
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
