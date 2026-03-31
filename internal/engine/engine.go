package engine

import (
	board "chessV2/internal/board"
	"chessV2/internal/movegen"
)

type Engine struct {
	moveGenerator   *movegen.PseudoLegalMoveGenerator
	positionUpdater board.MoveApplier
	usePerftTricks  bool
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

func NewEngine() *Engine {
	moveGenerator := movegen.NewPseudoLegalMoveGenerator()
	positionUpdater := board.NewPositionUpdater()

	return &Engine{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
		usePerftTricks:  true,
	}
}

func (e *Engine) SetPerftTricks(enabled bool) {
	e.usePerftTricks = enabled
	if enabled {
		e.positionUpdater = board.NewPositionUpdater()
		return
	}
	e.positionUpdater = board.NewPlainPositionUpdater()
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

func (e *Engine) LegalMoves(pos *board.Position) []board.Move {
	var buf [MaxLegalMoves]board.Move
	count := e.moveGenerator.LegalMovesInto(pos, e.positionUpdater, buf[:])
	moves := make([]board.Move, count)
	copy(moves, buf[:count])
	return moves
}

func (e *Engine) PerftDivide(pos *board.Position, depth int) (map[string]uint64, uint64) {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]board.Move
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
	res := make(map[string]uint64)
	total := uint64(0)

	rootMoves := moveBuffers[0][:]
	rootCount := e.moveGenerator.LegalMovesInto(pos, e.positionUpdater, rootMoves)
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

func (e *Engine) MoveGenerationTest(pos *board.Position, depth int) uint64 {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]board.Move
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
	return e.moveGenerationTestWithBuffers(pos, depth, 0, &moveBuffers, tt)
}

func (e *Engine) moveGenerationTestWithBuffers(pos *board.Position, depth int, ply int, moveBuffers *[MaxPerftPly][MaxLegalMoves]board.Move, tt *perftTT) uint64 {
	if depth == 1 {
		return uint64(1)
	}
	if tt != nil {
		if count, ok := tt.probe(pos.ZobristKey(), int8(depth)); ok {
			return count
		}
	}
	if e.usePerftTricks && depth == 2 {
		result := uint64(e.moveGenerator.LegalMovesInto(pos, e.positionUpdater, moveBuffers[ply][:]))
		if tt != nil {
			tt.store(pos.ZobristKey(), int8(depth), result)
		}
		return result
	}

	moves := moveBuffers[ply][:]
	moveCount := e.moveGenerator.LegalMovesInto(pos, e.positionUpdater, moves)
	posCount := uint64(0)
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := e.positionUpdater.MakeMove(pos, move)

		nextDepth := depth - 1
		nextDepthResult := e.moveGenerationTestWithBuffers(pos, nextDepth, ply+1, moveBuffers, tt)
		e.positionUpdater.UnMakeMove(pos, history)

		posCount += nextDepthResult
	}

	if tt != nil {
		tt.store(pos.ZobristKey(), int8(depth), posCount)
	}
	return posCount
}
