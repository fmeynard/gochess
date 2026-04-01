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
	Nodes           uint64
	QuiescenceNodes uint64
	Cutoffs         uint64
	Depth           int
	Time            time.Duration
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

	var (
		result Result
		err    error
	)
	if limits.MoveTime > 0 {
		result, err = s.searchIterative(pos, limits)
	} else {
		result, err = s.searchDepth(pos, limits.Depth, time.Time{}, limits.Stop, newRepetitionTracker(pos, limits.History))
	}
	if err != nil {
		return Result{}, err
	}
	return s.ensureBestMove(pos, result), nil
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
		iterPos := pos
		if refreshed, refreshErr := board.NewPositionFromFEN(pos.FEN()); refreshErr == nil {
			iterPos = refreshed
		}

		result, err := s.searchDepth(iterPos, depth, deadline, limits.Stop, newRepetitionTracker(iterPos, limits.History))
		if err != nil {
			if errors.Is(err, errSearchTimeout) || errors.Is(err, errSearchStopped) {
				if haveComplete {
					lastComplete.Stats.Time = time.Since(start)
					return lastComplete, nil
				}

				var fallbackMoves [256]board.Move
				moveCount := s.moveGenerator.LegalMovesInto(iterPos, s.positionUpdater, fallbackMoves[:])
				if moveCount == 0 {
					return Result{
						Score: terminalScore(iterPos, 0),
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
	orderMoves(pos, moves[:moveCount])

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
		return s.quiescence(pos, ply, alpha, beta, stats, deadline, stop, repetitions)
	}

	orderMoves(pos, moves[:moveCount])

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
			stats.Cutoffs++
			break
		}
	}

	return bestScore, nil
}

func (s *AlphaBetaSearcher) quiescence(pos *board.Position, ply int, alpha eval.Score, beta eval.Score, stats *Stats, deadline time.Time, stop <-chan struct{}, repetitions *repetitionTracker) (eval.Score, error) {
	if err := shouldStop(deadline, stop); err != nil {
		return 0, err
	}

	stats.QuiescenceNodes++
	if repetitions.isThreefold() {
		return eval.DrawScore, nil
	}

	standPat := s.evaluator.Evaluate(pos)
	if standPat >= beta {
		stats.Cutoffs++
		return beta, nil
	}
	if standPat > alpha {
		alpha = standPat
	}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(pos, s.positionUpdater, moves[:])
	if moveCount == 0 {
		return terminalScore(pos, ply), nil
	}

	orderMoves(pos, moves[:moveCount])
	for i := 0; i < moveCount; i++ {
		move := moves[i]
		if !isTacticalMove(pos, move) {
			continue
		}

		history := s.positionUpdater.MakeMove(pos, move)
		repetitions.push(pos.ZobristKey())
		score, err := s.quiescence(pos, ply+1, -beta, -alpha, stats, deadline, stop, repetitions)
		repetitions.pop()
		s.positionUpdater.UnMakeMove(pos, history)
		if err != nil {
			return 0, err
		}
		score = -score

		if score >= beta {
			stats.Cutoffs++
			return beta, nil
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha, nil
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

func orderMoves(pos *board.Position, moves []board.Move) {
	for i := 1; i < len(moves); i++ {
		move := moves[i]
		score := scoreMove(pos, move)
		j := i - 1
		for ; j >= 0 && score > scoreMove(pos, moves[j]); j-- {
			moves[j+1] = moves[j]
		}
		moves[j+1] = move
	}
}

func scoreMove(pos *board.Position, move board.Move) int {
	score := 0
	if isCaptureMove(pos, move) {
		captured := capturedPiece(pos, move)
		attacker := move.Piece().Type()
		score += 100000 + 10*pieceOrderValue(captured.Type()) - pieceOrderValue(attacker)
	}

	switch move.Flag() {
	case board.QueenPromotion:
		score += 50000 + pieceOrderValue(board.Queen)
	case board.RookPromotion:
		score += 50000 + pieceOrderValue(board.Rook)
	case board.BishopPromotion:
		score += 50000 + pieceOrderValue(board.Bishop)
	case board.KnightPromotion:
		score += 50000 + pieceOrderValue(board.Knight)
	}

	return score
}

func isTacticalMove(pos *board.Position, move board.Move) bool {
	if isCaptureMove(pos, move) {
		return true
	}

	switch move.Flag() {
	case board.QueenPromotion, board.RookPromotion, board.BishopPromotion, board.KnightPromotion:
		return true
	default:
		return false
	}
}

func isCaptureMove(pos *board.Position, move board.Move) bool {
	if move.Flag() == board.Capture || move.Flag() == board.EnPassant {
		return true
	}
	return pos.PieceAt(move.EndIdx()) != board.NoPiece
}

func capturedPiece(pos *board.Position, move board.Move) board.Piece {
	if move.Flag() == board.EnPassant {
		end := move.EndIdx()
		if move.Piece().IsWhite() {
			return pos.PieceAt(end - 8)
		}
		return pos.PieceAt(end + 8)
	}
	return pos.PieceAt(move.EndIdx())
}

func pieceOrderValue(pieceType int8) int {
	switch pieceType {
	case board.Pawn:
		return 100
	case board.Knight:
		return 320
	case board.Bishop:
		return 330
	case board.Rook:
		return 500
	case board.Queen:
		return 900
	case board.King:
		return 10000
	default:
		return 0
	}
}

func (s *AlphaBetaSearcher) ensureBestMove(pos *board.Position, result Result) Result {
	if result.BestMove != (board.Move{}) {
		return result
	}

	root := pos
	if refreshed, err := board.NewPositionFromFEN(pos.FEN()); err == nil {
		root = refreshed
	}

	var moves [256]board.Move
	moveCount := s.moveGenerator.LegalMovesInto(root, s.positionUpdater, moves[:])
	if moveCount > 0 {
		result.BestMove = moves[0]
	}
	return result
}
