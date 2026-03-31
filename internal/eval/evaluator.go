package eval

import board "chessV2/internal/board"

type Evaluator interface {
	Evaluate(pos *board.Position) Score
}

type ZeroEvaluator struct{}

func NewZeroEvaluator() *ZeroEvaluator {
	return &ZeroEvaluator{}
}

func (e *ZeroEvaluator) Evaluate(pos *board.Position) Score {
	return DrawScore
}

