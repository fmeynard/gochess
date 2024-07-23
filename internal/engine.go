package internal

import "math/bits"

type Engine struct {
	moveGenerator   IMoveGenerator
	positionUpdater IPositionUpdater
}

type IPositionUpdater interface {
	MakeMove(pos *Position, move Move) *MoveHistory
	UnMakeMove(pos *Position, history *MoveHistory)
	IsMoveAffectsKing(pos *Position, m Move, kingColor int8) bool
}

type IMoveGenerator interface {
	PawnPseudoLegalMoves(pos *Position, idx int8) ([]int8, int8)
	SliderPseudoLegalMoves(pos *Position, idx int8, pieceType int8) []int8
	KingPseudoLegalMoves(pos *Position, idx int8) []int8
	KnightPseudoLegalMoves(pos *Position, idx int8) []int8
}

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

func (e *Engine) LegalMoves(pos *Position) []Move {
	var moves []Move

	piecesMask := pos.OccupancyMask(pos.activeColor)
	for piecesMask != 0 { //for idx := int8(0); idx < 64; idx++ {
		idx := int8(bits.TrailingZeros64(piecesMask))
		//piece := pos.PieceAt(idx)

		pieceMask := uint64(1 << idx)
		var (
			pseudoLegalMoves []int8
			pieceType        int8
		)

		switch {
		case pieceMask&pos.pawnBoard != 0:
			pseudoLegalMoves, _ = e.moveGenerator.PawnPseudoLegalMoves(pos, idx)
			pieceType = Pawn
		case pieceMask&pos.knightBoard != 0:
			pseudoLegalMoves = e.moveGenerator.KnightPseudoLegalMoves(pos, idx)
			pieceType = Knight
		case pieceMask&pos.bishopBoard != 0:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Bishop)
			pieceType = Bishop
		case pieceMask&pos.rookBoard != 0:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Rook)
			pieceType = Rook
		case pieceMask&pos.queenBoard != 0:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Queen)
			pieceType = Queen
		case pieceMask&pos.kingBoard != 0:
			pseudoLegalMoves = e.moveGenerator.KingPseudoLegalMoves(pos, idx)
			pieceType = King
		}

		for _, pseudoLegalMoveIdx := range pseudoLegalMoves {
			initialColor := pos.activeColor
			pseudoLegalMove := NewMove(Piece(pos.activeColor|pieceType), idx, pseudoLegalMoveIdx, NormalMove)

			// skip further checks if not needed
			if !pos.IsCheck() && !e.positionUpdater.IsMoveAffectsKing(pos, pseudoLegalMove, pos.activeColor) {
				moves = append(moves, pseudoLegalMove)
				continue
			}

			history := e.positionUpdater.MakeMove(pos, pseudoLegalMove)

			if !IsKingInCheck(pos, initialColor) { // check is the new position that the initial color is not in check
				moves = append(moves, pseudoLegalMove)
			}

			e.positionUpdater.UnMakeMove(pos, history)
			history = nil
		}

		piecesMask &^= 1 << idx
	}

	return moves
}

func (e *Engine) PerftDivide(pos *Position, depth int) (map[string]uint64, uint64) {
	res := make(map[string]uint64)
	total := uint64(0)

	for _, move := range e.LegalMoves(pos) {
		history := e.positionUpdater.MakeMove(pos, move)
		res[move.UCI()] = e.MoveGenerationTest(pos, depth)
		total += res[move.UCI()]
		e.positionUpdater.UnMakeMove(pos, history)
		history = nil
	}

	return res, total
}

func (e *Engine) MoveGenerationTest(pos *Position, depth int) uint64 {
	if depth == 1 {
		return uint64(1)
	}

	posCount := uint64(0)
	for _, move := range e.LegalMoves(pos) {
		history := e.positionUpdater.MakeMove(pos, move)

		nextDepth := depth - 1
		nextDepthResult := e.MoveGenerationTest(pos, nextDepth)
		e.positionUpdater.UnMakeMove(pos, history)

		posCount += nextDepthResult
	}

	return posCount
}
