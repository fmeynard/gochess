package chess

type Engine struct {
	moveGenerator   *PseudoLegalMoveGenerator
	positionUpdater moveApplier
	usePerftTricks  bool
}

const (
	MaxPerftPly   = 64
	MaxLegalMoves = 256
	MaxTargets    = 28
)

func NewEngine() *Engine {
	moveGenerator := NewPseudoLegalMoveGenerator()
	positionUpdater := NewPositionUpdater(moveGenerator)

	return &Engine{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
		usePerftTricks:  true,
	}
}

func (e *Engine) SetPerftTricks(enabled bool) {
	e.usePerftTricks = enabled
	if enabled {
		e.positionUpdater = NewPositionUpdater(e.moveGenerator)
		return
	}
	e.positionUpdater = NewPlainPositionUpdater(e.moveGenerator)
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

func (e *Engine) LegalMoves(pos *Position) []Move {
	var buf [MaxLegalMoves]Move
	count := e.legalMovesInto(pos, buf[:])
	moves := make([]Move, count)
	copy(moves, buf[:count])
	return moves
}

func (e *Engine) PerftDivide(pos *Position, depth int) (map[string]uint64, uint64) {
	var moveBuffers [MaxPerftPly][MaxLegalMoves]Move
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
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
	var tt *perftTT
	if e.usePerftTricks {
		tt = newPerftTT()
	}
	return e.moveGenerationTestWithBuffers(pos, depth, 0, &moveBuffers, tt)
}

func (e *Engine) moveGenerationTestWithBuffers(pos *Position, depth int, ply int, moveBuffers *[MaxPerftPly][MaxLegalMoves]Move, tt *perftTT) uint64 {
	if depth == 1 {
		return uint64(1)
	}
	if tt != nil {
		if count, ok := tt.probe(pos.zobristKey, int8(depth)); ok {
			return count
		}
	}
	if e.usePerftTricks && depth == 2 {
		result := uint64(e.legalMovesInto(pos, moveBuffers[ply][:]))
		if tt != nil {
			tt.store(pos.zobristKey, int8(depth), result)
		}
		return result
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

	if tt != nil {
		tt.store(pos.zobristKey, int8(depth), posCount)
	}
	return posCount
}
