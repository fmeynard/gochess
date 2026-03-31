package internal

import "math/bits"

type Engine struct {
	moveGenerator   IMoveGenerator
	positionUpdater IPositionUpdater
}

type IPositionUpdater interface {
	MakeMove(pos *Position, move Move) MoveHistory
	UnMakeMove(pos *Position, history MoveHistory)
	IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool
}

type IMoveGenerator interface {
	PawnPseudoLegalMoves(pos *Position, idx int8) ([]int8, int8)
	SliderPseudoLegalMoves(pos *Position, idx int8, pieceType int8) []int8
	KingPseudoLegalMoves(pos *Position, idx int8) []int8
	KnightPseudoLegalMoves(pos *Position, idx int8) []int8
	PawnPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) (int, int8)
	SliderPseudoLegalMovesInto(pos *Position, idx int8, pieceType int8, dst []int8) int
	KingPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) int
	KnightPseudoLegalMovesInto(pos *Position, idx int8, dst []int8) int
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

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

func (e *Engine) legalMovesInto(pos *Position, dst []Move) int {
	count := 0
	var targets [MaxTargets]int8
	positionCheck := pos.IsCheck()

	piecesMask := pos.OccupancyMask(pos.activeColor)
	for piecesMask != 0 {
		idx := int8(bits.TrailingZeros64(piecesMask))

		pieceMask := uint64(1 << idx)
		var (
			pseudoCount  int
			pieceType    int8
			piece        Piece
			initialColor = pos.activeColor
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
		for i := 0; i < pseudoCount; i++ {
			targetIdx := targets[i]
			if pieceType == Pawn && isPromotionSquare(pos.activeColor, targetIdx) {
				flags := [4]int8{QueenPromotion, KnightPromotion, BishopPromotion, RookPromotion}
				for _, flag := range flags {
					move := NewMove(piece, idx, targetIdx, flag)
					if e.isCastleMove(pieceType, move) && !e.isCastlePathSafe(pos, move) {
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
			if e.isCastleMove(pieceType, move) && !e.isCastlePathSafe(pos, move) {
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
	res := make(map[string]uint64)
	total := uint64(0)

	rootMoves := moveBuffers[0][:]
	rootCount := e.legalMovesInto(pos, rootMoves)
	for i := 0; i < rootCount; i++ {
		move := rootMoves[i]
		history := e.positionUpdater.MakeMove(pos, move)
		res[move.UCI()] = e.moveGenerationTestWithBuffers(pos, depth, 1, &moveBuffers)
		total += res[move.UCI()]
		e.positionUpdater.UnMakeMove(pos, history)
	}

	return res, total
}

func (e *Engine) MoveGenerationTest(pos *Position, depth int) uint64 {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	return e.moveGenerationTestWithBuffers(pos, depth, 0, &moveBuffers)
}

func (e *Engine) moveGenerationTestWithBuffers(pos *Position, depth int, ply int, moveBuffers *[MaxPerftPly][MaxLegalMoves]Move) uint64 {
	if depth == 1 {
		return uint64(1)
	}

	moves := moveBuffers[ply][:]
	moveCount := e.legalMovesInto(pos, moves)
	posCount := uint64(0)
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := e.positionUpdater.MakeMove(pos, move)

		nextDepth := depth - 1
		nextDepthResult := e.moveGenerationTestWithBuffers(pos, nextDepth, ply+1, moveBuffers)
		e.positionUpdater.UnMakeMove(pos, history)

		posCount += nextDepthResult
	}

	return posCount
}
