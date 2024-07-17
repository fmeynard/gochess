package internal

type Engine struct {
	game            *Game
	moveGenerator   IMoveGenerator
	positionUpdater IPositionUpdater
}

type IPositionUpdater interface {
	MakeMove(initPos *Position, move Move) MoveHistory
	UnMakeMove(initPos *Position, move Move, history MoveHistory)
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

	for idx := int8(0); idx < 64; idx++ {
		piece := pos.PieceAt(idx)
		if piece.Color() != pos.activeColor {
			continue
		}

		var pseudoLegalMoves []int8

		switch piece.Type() {
		case Pawn:
			pseudoLegalMoves, _ = e.moveGenerator.PawnPseudoLegalMoves(pos, idx)
		case Rook:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Rook)
		case Bishop:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Bishop)
		case Knight:
			pseudoLegalMoves = e.moveGenerator.KnightPseudoLegalMoves(pos, idx)
		case Queen:
			pseudoLegalMoves = e.moveGenerator.SliderPseudoLegalMoves(pos, idx, Queen)
		case King:
			pseudoLegalMoves = e.moveGenerator.KingPseudoLegalMoves(pos, idx)
		}

		for _, pseudoLegalMoveIdx := range pseudoLegalMoves {
			initialColor := pos.activeColor
			pseudoLegalMove := NewMove(piece, idx, pseudoLegalMoveIdx, NormalMove)
			history := e.positionUpdater.MakeMove(pos, pseudoLegalMove)
			if !IsKingInCheck(pos, initialColor) { // check is the new position that the initial color is not in check
				moves = append(moves, pseudoLegalMove)
			}
			e.positionUpdater.UnMakeMove(pos, pseudoLegalMove, history)
		}
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
		e.positionUpdater.UnMakeMove(pos, move, history)
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
		e.positionUpdater.UnMakeMove(pos, move, history)

		posCount += nextDepthResult
	}

	return posCount
}
