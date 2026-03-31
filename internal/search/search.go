package search

import (
	board "chessV2/internal/board"
	"chessV2/internal/eval"
	"chessV2/internal/movegen"
	"errors"
	"time"
)

var ErrInvalidLimits = errors.New("invalid search limits")
var errSearchTimeout = errors.New("search timeout")
var errSearchStopped = errors.New("search stopped")

type Limits struct {
	Depth    int
	MoveTime time.Duration
	Stop     <-chan struct{}
	History  []uint64
}

type Stats struct {
	Nodes uint64
	Depth int
	Time  time.Duration
}

type Result struct {
	BestMove board.Move
	Score    eval.Score
	Stats    Stats
}

type Searcher interface {
	Search(pos *board.Position, limits Limits) (Result, error)
	NewGame()
}

type AlphaBetaSearcher struct {
	moveGenerator   *movegen.PseudoLegalMoveGenerator
	positionUpdater board.MoveApplier
	evaluator       eval.Evaluator
}

type repetitionTracker struct {
	stack  []uint64
	counts map[uint64]uint8
}

func NewAlphaBetaSearcher(moveGenerator *movegen.PseudoLegalMoveGenerator, positionUpdater board.MoveApplier, evaluator eval.Evaluator) *AlphaBetaSearcher {
	return &AlphaBetaSearcher{
		moveGenerator:   moveGenerator,
		positionUpdater: positionUpdater,
		evaluator:       evaluator,
	}
}

func (s *AlphaBetaSearcher) Search(pos *board.Position, limits Limits) (Result, error) {
	if limits.Depth <= 0 && limits.MoveTime <= 0 {
		return Result{}, ErrInvalidLimits
	}

	if limits.MoveTime > 0 {
		return s.searchIterative(pos, limits)
	}
	return s.searchDepth(pos, limits.Depth, time.Time{}, limits.Stop, newRepetitionTracker(pos, limits.History))
}

func (s *AlphaBetaSearcher) searchIterative(pos *board.Position, limits Limits) (Result, error) {
	start := time.Now()
	deadline := start.Add(limits.MoveTime)
	maxDepth := limits.Depth
	if maxDepth <= 0 {
		maxDepth = 64
	}

	var lastComplete Result
	var haveComplete bool

	for depth := 1; depth <= maxDepth; depth++ {
		result, err := s.searchDepth(pos, depth, deadline, limits.Stop, newRepetitionTracker(pos, limits.History))
		if err != nil {
			if errors.Is(err, errSearchTimeout) || errors.Is(err, errSearchStopped) {
				if haveComplete {
					lastComplete.Stats.Time = time.Since(start)
					return lastComplete, nil
				}

				var fallbackMoves [256]board.Move
				moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, fallbackMoves[:])
				if moveCount == 0 {
					return Result{
						Score: terminalScore(pos, 0),
						Stats: Stats{
							Nodes: 1,
							Time:  time.Since(start),
						},
					}, nil
				}

				return Result{
					BestMove: fallbackMoves[0],
					Score:    eval.DrawScore,
					Stats: Stats{
						Depth: 0,
						Time:  time.Since(start),
					},
				}, nil
			}
			return Result{}, err
		}

		lastComplete = result
		haveComplete = true
		if time.Now().After(deadline) {
			break
		}
	}

	lastComplete.Stats.Time = time.Since(start)
	return lastComplete, nil
}

func (s *AlphaBetaSearcher) searchDepth(pos *board.Position, depth int, deadline time.Time, stop <-chan struct{}, repetitions *repetitionTracker) (Result, error) {
	if depth <= 0 {
		return Result{}, ErrInvalidLimits
	}

	start := time.Now()
	stats := Stats{Depth: depth}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return Result{
			Score: terminalScore(pos, 0),
			Stats: Stats{
				Nodes: 1,
				Depth: depth,
				Time:  time.Since(start),
			},
		}, nil
	}

	bestMove := moves[0]
	bestScore := -eval.InfinityScore
	alpha := -eval.InfinityScore
	beta := eval.InfinityScore
	haveComplete := false

	for i := 0; i < moveCount; i++ {
		if err := shouldStop(deadline, stop); err != nil {
			if haveComplete {
				stats.Time = time.Since(start)
				return Result{
					BestMove: bestMove,
					Score:    bestScore,
					Stats:    stats,
				}, nil
			}
			return Result{
				BestMove: moves[0],
				Score:    eval.DrawScore,
				Stats: Stats{
					Depth: depth,
					Time:  time.Since(start),
				},
			}, nil
		}

		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		repetitions.push(pos.ZobristKey())
		score, err := s.negamax(pos, depth-1, 1, -beta, -alpha, &stats, deadline, stop, repetitions)
		repetitions.pop()
		s.positionUpdater.UnMakeMove(pos, history)
		if err != nil {
			if (errors.Is(err, errSearchTimeout) || errors.Is(err, errSearchStopped)) && haveComplete {
				stats.Time = time.Since(start)
				return Result{
					BestMove: bestMove,
					Score:    bestScore,
					Stats:    stats,
				}, nil
			}
			if errors.Is(err, errSearchTimeout) || errors.Is(err, errSearchStopped) {
				return Result{
					BestMove: moves[0],
					Score:    eval.DrawScore,
					Stats: Stats{
						Depth: depth,
						Time:  time.Since(start),
					},
				}, nil
			}
			return Result{}, err
		}
		score = -score

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		haveComplete = true
		if score > alpha {
			alpha = score
		}
	}

	stats.Time = time.Since(start)
	return Result{
		BestMove: bestMove,
		Score:    bestScore,
		Stats:    stats,
	}, nil
}

func (s *AlphaBetaSearcher) NewGame() {}

func (s *AlphaBetaSearcher) negamax(pos *board.Position, depth int, ply int, alpha eval.Score, beta eval.Score, stats *Stats, deadline time.Time, stop <-chan struct{}, repetitions *repetitionTracker) (eval.Score, error) {
	if err := shouldStop(deadline, stop); err != nil {
		return 0, err
	}

	stats.Nodes++
	if repetitions.isThreefold() {
		return eval.DrawScore, nil
	}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return terminalScore(pos, ply), nil
	}

	if depth == 0 {
		return s.evaluator.Evaluate(pos), nil
	}

	bestScore := -eval.InfinityScore
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		history := s.positionUpdater.MakeMove(pos, move)
		repetitions.push(pos.ZobristKey())
		score, err := s.negamax(pos, depth-1, ply+1, -beta, -alpha, stats, deadline, stop, repetitions)
		repetitions.pop()
		s.positionUpdater.UnMakeMove(pos, history)
		if err != nil {
			return 0, err
		}
		score = -score

		if score > bestScore {
			bestScore = score
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}

	return bestScore, nil
}

func terminalScore(pos *board.Position, ply int) eval.Score {
	if movegen.IsKingInCheck(pos, pos.ActiveColor()) {
		return eval.MatedIn(ply)
	}
	return eval.DrawScore
}

func shouldStop(deadline time.Time, stop <-chan struct{}) error {
	select {
	case <-stop:
		return errSearchStopped
	default:
	}

	if !deadline.IsZero() && time.Now().After(deadline) {
		return errSearchTimeout
	}

	return nil
}

func newRepetitionTracker(pos *board.Position, history []uint64) *repetitionTracker {
	tracker := &repetitionTracker{
		stack:  make([]uint64, 0, len(history)+1),
		counts: make(map[uint64]uint8, len(history)+1),
	}

	for _, key := range history {
		tracker.push(key)
	}

	if len(tracker.stack) == 0 || tracker.stack[len(tracker.stack)-1] != pos.ZobristKey() {
		tracker.push(pos.ZobristKey())
	}

	return tracker
}

func (t *repetitionTracker) push(key uint64) {
	t.stack = append(t.stack, key)
	t.counts[key]++
}

func (t *repetitionTracker) pop() {
	lastIdx := len(t.stack) - 1
	key := t.stack[lastIdx]
	t.stack = t.stack[:lastIdx]
	if t.counts[key] <= 1 {
		delete(t.counts, key)
		return
	}
	t.counts[key]--
}

func (t *repetitionTracker) isThreefold() bool {
	if len(t.stack) == 0 {
		return false
	}
	return t.counts[t.stack[len(t.stack)-1]] >= 3
}
