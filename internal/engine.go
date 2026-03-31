package internal

import "math/bits"

type Engine struct {
	moveGenerator   *BitsBoardMoveGenerator
	positionUpdater *PositionUpdater
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

var promotionFlags = [4]int8{QueenPromotion, KnightPromotion, BishopPromotion, RookPromotion}

var rayDirections = [8]int8{West, East, South, North, SouthWest, SouthEast, NorthWest, NorthEast}

func NewEngine() *Engine {
	moveGenerator := NewBitsBoardMoveGenerator()
	positionUpdater := NewPositionUpdater(moveGenerator)

	return &Engine{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
	}
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

func (e *Engine) isCastleMove(pieceType int8, move Move) bool {
	return pieceType == King && absInt8(move.EndIdx()-move.StartIdx()) == 2
}

func (e *Engine) isCastlePathSafe(pos *Position, move Move) bool {
	initialColor := pos.activeColor
	if IsKingInCheck(pos, initialColor) {
		return false
	}

	step := int8(1)
	if move.EndIdx() < move.StartIdx() {
		step = -1
	}

	intermediateMove := NewMove(Piece(initialColor|King), move.StartIdx(), move.StartIdx()+step, NormalMove)
	history := e.positionUpdater.MakeMove(pos, intermediateMove)
	isIntermediateSquareSafe := !IsKingInCheck(pos, initialColor)
	e.positionUpdater.UnMakeMove(pos, history)

	return isIntermediateSquareSafe
}

func computePins(pos *Position, kingIdx int8, friendlyOcc, enemyOcc uint64) (uint64, [64]uint64) {
	var (
		pinnedMask uint64
		pinRays    [64]uint64
	)

	rookQueens := (pos.rookBoard | pos.queenBoard) & enemyOcc
	bishopQueens := (pos.bishopBoard | pos.queenBoard) & enemyOcc

	for dirIdx := 0; dirIdx < 8; dirIdx++ {
		var sliders uint64
		if dirIdx < 4 {
			sliders = rookQueens
		} else {
			sliders = bishopQueens
		}
		if sliders == 0 {
			continue
		}

		ray := sliderAttackMasks[kingIdx][dirIdx]
		if ray == 0 {
			continue
		}

		dir := rayDirections[dirIdx]
		first := firstBlockerOnRay(pos.occupied, ray, dir)
		if first == 0 || first&friendlyOcc == 0 {
			continue
		}

		second := firstBlockerOnRay(pos.occupied&^first, ray, dir)
		if second == 0 || second&sliders == 0 {
			continue
		}

		pinnedMask |= first
		pinRays[bits.TrailingZeros64(first)] = ray
	}

	return pinnedMask, pinRays
}

func (e *Engine) legalMovesInto(pos *Position, dst []Move) int {
	count := 0
	var targets [MaxTargets]int8
	positionCheck := pos.IsCheck()
	initialColor := pos.activeColor

	var (
		pinnedMask uint64
		pinRays    [64]uint64
		kingIdx    int8
	)
	if pos.activeColor == White {
		kingIdx = pos.whiteKingIdx
	} else {
		kingIdx = pos.blackKingIdx
	}
	if !positionCheck {
		pinnedMask, pinRays = computePins(pos, kingIdx, pos.OccupancyMask(pos.activeColor), pos.OpponentOccupiedMask())
	}

	piecesMask := pos.OccupancyMask(pos.activeColor)
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))

		pieceMask := uint64(1 << idx)
		var (
			pseudoCount int
			pieceType   int8
			piece       Piece
		)

		switch {
		case pieceMask&pos.pawnBoard != 0:
			pseudoCount, _ = e.moveGenerator.PawnPseudoLegalMovesInto(pos, idx, targets[:])
			pieceType = Pawn
		case pieceMask&pos.knightBoard != 0:
			pseudoCount = e.moveGenerator.KnightPseudoLegalMovesInto(pos, idx, targets[:])
			pieceType = Knight
		case pieceMask&pos.bishopBoard != 0:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Bishop, targets[:])
			pieceType = Bishop
		case pieceMask&pos.rookBoard != 0:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Rook, targets[:])
			pieceType = Rook
		case pieceMask&pos.queenBoard != 0:
			pseudoCount = e.moveGenerator.SliderPseudoLegalMovesInto(pos, idx, Queen, targets[:])
			pieceType = Queen
		case pieceMask&pos.kingBoard != 0:
			pseudoCount = e.moveGenerator.KingPseudoLegalMovesInto(pos, idx, targets[:])
			pieceType = King
		}

		piece = Piece(pos.activeColor | pieceType)
		isPinned := !positionCheck && pinnedMask&pieceMask != 0

		for i := 0; i < pseudoCount; i++ {
			targetIdx := targets[i]
			if pieceType == Pawn && isPromotionSquare(pos.activeColor, targetIdx) {
				for _, flag := range promotionFlags {
					move := NewMove(piece, idx, targetIdx, flag)
					if !positionCheck && pieceType != King {
						if !isPinned {
							dst[count] = move
							count++
							continue
						}
						if pinRays[idx]&(1<<targetIdx) != 0 {
							dst[count] = move
							count++
							continue
						}
						continue
					}
					if !positionCheck && !e.positionUpdater.IsMoveAffectsKing(pos, move, pos.activeColor) {
						dst[count] = move
						count++
						continue
					}
					history := e.positionUpdater.MakeMove(pos, move)
					if !IsKingInCheck(pos, initialColor) {
						dst[count] = move
						count++
					}
					e.positionUpdater.UnMakeMove(pos, history)
				}
				continue
			}

			move := NewMove(piece, idx, targetIdx, NormalMove)
			if pieceType == King && absInt8(targetIdx-idx) == 2 {
				if !e.isCastlePathSafe(pos, move) {
					continue
				}
			}

			if !positionCheck && pieceType != King {
				isEP := pieceType == Pawn && pos.enPassantIdx != NoEnPassant && targetIdx == pos.enPassantIdx
				if !isEP {
					if !isPinned {
						dst[count] = move
						count++
						continue
					}
					if pinRays[idx]&(1<<targetIdx) != 0 {
						dst[count] = move
						count++
						continue
					}
					continue
				}
			}

			if !positionCheck && !e.positionUpdater.IsMoveAffectsKing(pos, move, pos.activeColor) {
				dst[count] = move
				count++
				continue
			}

			history := e.positionUpdater.MakeMove(pos, move)
			if !IsKingInCheck(pos, initialColor) {
				dst[count] = move
				count++
			}
			e.positionUpdater.UnMakeMove(pos, history)
		}

		piecesMask &^= 1 << idx
	}

	return count
}

func (e *Engine) LegalMoves(pos *Position) []Move {
	var buf [MaxLegalMoves]Move
	count := e.legalMovesInto(pos, buf[:])
	moves := make([]Move, count)
	copy(moves, buf[:count])
	return moves
}

func (e *Engine) PerftDivide(pos *Position, depth int) (map[string]uint64, uint64) {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	tt := newPerftTT()
	res := make(map[string]uint64)
	total := uint64(0)

	rootMoves := moveBuffers[0][:]
	rootCount := e.legalMovesInto(pos, rootMoves)
	for i := 0; i < rootCount; i++ {
		move := rootMoves[i]
		history := e.positionUpdater.MakeMove(pos, move)
		uci := move.UCI()
		res[uci] = e.moveGenerationTestWithBuffers(pos, depth, 1, &moveBuffers, tt)
		total += res[uci]
		e.positionUpdater.UnMakeMove(pos, history)
	}

	return res, total
}

func (e *Engine) MoveGenerationTest(pos *Position, depth int) uint64 {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	tt := newPerftTT()
	return e.moveGenerationTestWithBuffers(pos, depth, 0, &moveBuffers, tt)
}

func (e *Engine) moveGenerationTestWithBuffers(pos *Position, depth int, ply int, moveBuffers *[MaxPerftPly][MaxLegalMoves]Move, tt *perftTT) uint64 {
	if depth == 1 {
		return uint64(1)
	}
	if depth == 2 {
		return uint64(e.legalMovesInto(pos, moveBuffers[ply][:]))
	}
	if count, ok := tt.probe(pos.zobristKey, int8(depth)); ok {
		return count
	}

	moves := moveBuffers[ply][:]
	moveCount := e.legalMovesInto(pos, moves)
	posCount := uint64(0)
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := e.positionUpdater.MakeMove(pos, move)

		nextDepth := depth - 1
		nextDepthResult := e.moveGenerationTestWithBuffers(pos, nextDepth, ply+1, moveBuffers, tt)
		e.positionUpdater.UnMakeMove(pos, history)

		posCount += nextDepthResult
	}

	tt.store(pos.zobristKey, int8(depth), posCount)
	return posCount
}
