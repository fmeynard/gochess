package uci

import (
	"bufio"
	board "chessV2/internal/board"
	"chessV2/internal/engine"
	"chessV2/internal/search"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	engineName   = "GoChess"
	engineAuthor = "fmeynard"
)

type Server struct {
	mu           sync.Mutex
	writeMu      sync.Mutex
	engine       *engine.Engine
	position     *board.Position
	positionKeys []uint64
	activeSearch *activeSearch
}

type activeSearch struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func NewServer(e *engine.Engine) (*Server, error) {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return nil, err
	}

	return &Server{
		engine:       e,
		position:     pos,
		positionKeys: []uint64{pos.ZobristKey()},
	}, nil
}

func (s *Server) Run(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		quit, err := s.handleCommand(line, out)
		if err != nil {
			fmt.Fprintf(out, "info string error %s\n", sanitizeInfo(err.Error()))
		}
		if quit {
			return nil
		}
	}

	return scanner.Err()
}

func (s *Server) handleCommand(line string, out io.Writer) (bool, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false, nil
	}

	switch fields[0] {
	case "uci":
		fmt.Fprintf(out, "id name %s\n", engineName)
		fmt.Fprintf(out, "id author %s\n", engineAuthor)
		fmt.Fprintln(out, "uciok")
	case "isready":
		s.stopSearch(true)
		fmt.Fprintln(out, "readyok")
	case "ucinewgame":
		s.stopSearch(true)
		s.engine.StartGame()
		return false, s.resetToStartPos()
	case "position":
		s.stopSearch(true)
		return false, s.handlePosition(fields[1:])
	case "go":
		return false, s.handleGo(fields[1:], out)
	case "stop":
		s.stopSearch(false)
		return false, nil
	case "quit":
		s.stopSearch(true)
		return true, nil
	}

	return false, nil
}

func (s *Server) handlePosition(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing position arguments")
	}

	var pos *board.Position
	var err error
	i := 0

	switch args[0] {
	case "startpos":
		pos, err = board.NewPositionFromFEN(board.FenStartPos)
		if err != nil {
			return err
		}
		i = 1
	case "fen":
		if len(args) < 7 {
			return fmt.Errorf("invalid fen position command")
		}
		fen := strings.Join(args[1:7], " ")
		pos, err = board.NewPositionFromFEN(fen)
		if err != nil {
			return err
		}
		i = 7
	default:
		return fmt.Errorf("unsupported position command")
	}

	if i < len(args) {
		if args[i] != "moves" {
			return fmt.Errorf("unexpected token in position command: %s", args[i])
		}
		positionKeys, err := s.engine.ApplyUCIMovesWithPositionKeys(pos, args[i+1:])
		if err != nil {
			return err
		}
		s.position = pos
		s.positionKeys = positionKeys
		return nil
	}

	s.position = pos
	s.positionKeys = []uint64{pos.ZobristKey()}
	return nil
}

func (s *Server) handleGo(args []string, out io.Writer) error {
	snapshot, history := s.searchSnapshot()
	limits, err := parseGoLimits(args, snapshot.ActiveColor())
	if err != nil {
		return err
	}

	s.stopSearch(true)
	active := &activeSearch{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	s.mu.Lock()
	s.activeSearch = active
	s.mu.Unlock()

	go s.runSearch(active, snapshot, history, limits, out)
	return nil
}

func (s *Server) resetToStartPos() error {
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return err
	}
	s.position = pos
	s.positionKeys = []uint64{pos.ZobristKey()}
	return nil
}

func (s *Server) searchSnapshot() (*board.Position, []uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := make([]uint64, len(s.positionKeys))
	copy(history, s.positionKeys)
	return s.position.Clone(), history
}

func (s *Server) runSearch(active *activeSearch, pos *board.Position, history []uint64, parsedLimits struct {
	Depth    int
	MoveTime time.Duration
}, out io.Writer) {
	defer close(active.done)

	result, err := s.engine.Search(pos, search.Limits{
		Depth:    parsedLimits.Depth,
		MoveTime: parsedLimits.MoveTime,
		Stop:     active.stop,
		History:  history,
	})

	s.mu.Lock()
	if s.activeSearch == active {
		s.activeSearch = nil
	}
	s.mu.Unlock()

	if err != nil {
		s.writef(out, "info string error %s\n", sanitizeInfo(err.Error()))
		return
	}

	result = s.ensureBestMove(pos, result)
	s.writeResult(out, result)
}

func (s *Server) stopSearch(wait bool) {
	s.mu.Lock()
	active := s.activeSearch
	s.mu.Unlock()

	if active == nil {
		return
	}

	active.once.Do(func() {
		close(active.stop)
	})

	if wait {
		<-active.done
	}
}

func (s *Server) writeResult(out io.Writer, result search.Result) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	writeInfo(out, adaptResult(result))
	bestMove := "0000"
	if result.BestMove != (board.Move{}) {
		bestMove = result.BestMove.UCI()
	}
	if bestMove == "" {
		bestMove = "0000"
	}
	fmt.Fprintf(out, "bestmove %s\n", bestMove)
}

func (s *Server) writef(out io.Writer, format string, args ...any) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	fmt.Fprintf(out, format, args...)
}

func parseGoLimits(args []string, activeColor int8) (limits struct {
	Depth    int
	MoveTime time.Duration
}, err error) {
	var (
		whiteTime  time.Duration
		blackTime  time.Duration
		whiteInc   time.Duration
		blackInc   time.Duration
		movesToGo  int
		haveClocks bool
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "depth":
			if i+1 >= len(args) {
				return limits, fmt.Errorf("missing go depth value")
			}
			depth, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return limits, fmt.Errorf("invalid go depth value")
			}
			limits.Depth = depth
			i++
		case "movetime":
			if i+1 >= len(args) {
				return limits, fmt.Errorf("missing go movetime value")
			}
			ms, convErr := strconv.Atoi(args[i+1])
			if convErr != nil {
				return limits, fmt.Errorf("invalid go movetime value")
			}
			limits.MoveTime = time.Duration(ms) * time.Millisecond
			i++
		case "wtime":
			whiteTime, err = parseGoDurationArg(args, i+1, "wtime")
			if err != nil {
				return limits, err
			}
			haveClocks = true
			i++
		case "btime":
			blackTime, err = parseGoDurationArg(args, i+1, "btime")
			if err != nil {
				return limits, err
			}
			haveClocks = true
			i++
		case "winc":
			whiteInc, err = parseGoDurationArg(args, i+1, "winc")
			if err != nil {
				return limits, err
			}
			i++
		case "binc":
			blackInc, err = parseGoDurationArg(args, i+1, "binc")
			if err != nil {
				return limits, err
			}
			i++
		case "movestogo":
			if i+1 >= len(args) {
				return limits, fmt.Errorf("missing go movestogo value")
			}
			movesToGo, err = strconv.Atoi(args[i+1])
			if err != nil || movesToGo < 0 {
				return limits, fmt.Errorf("invalid go movestogo value")
			}
			i++
		}
	}

	if limits.MoveTime <= 0 && haveClocks {
		limits.MoveTime = allocateClockMoveTime(activeColor, whiteTime, blackTime, whiteInc, blackInc, movesToGo)
	}

	if limits.MoveTime <= 0 && limits.Depth <= 0 {
		limits.Depth = 1
	}

	return limits, nil
}

func parseGoDurationArg(args []string, valueIndex int, name string) (time.Duration, error) {
	if valueIndex >= len(args) {
		return 0, fmt.Errorf("missing go %s value", name)
	}

	ms, err := strconv.Atoi(args[valueIndex])
	if err != nil || ms < 0 {
		return 0, fmt.Errorf("invalid go %s value", name)
	}

	return time.Duration(ms) * time.Millisecond, nil
}

func allocateClockMoveTime(activeColor int8, whiteTime time.Duration, blackTime time.Duration, whiteInc time.Duration, blackInc time.Duration, movesToGo int) time.Duration {
	remaining := whiteTime
	increment := whiteInc
	if activeColor == board.Black {
		remaining = blackTime
		increment = blackInc
	}

	if remaining <= 0 {
		return 0
	}

	if movesToGo <= 0 {
		movesToGo = 30
	}

	const moveOverhead = 50 * time.Millisecond

	reserve := remaining / 20
	if reserve < moveOverhead {
		reserve = moveOverhead
	}
	maxBudget := remaining - reserve
	if maxBudget <= 0 {
		if remaining > moveOverhead {
			return remaining - moveOverhead
		}
		return time.Millisecond
	}

	budget := remaining/time.Duration(movesToGo) + (increment * 3 / 4)
	if budget <= 0 {
		budget = time.Millisecond
	}
	if budget > maxBudget {
		budget = maxBudget
	}
	if budget < time.Millisecond {
		return time.Millisecond
	}

	return budget
}

func writeInfo(out io.Writer, result searchResultLike) {
	timeMs := result.searchTime().Milliseconds()
	score := result.searchScore()
	if score > 29000 || score < -29000 {
		matePly := int((30000 - absScore(score) + 1) / 2)
		if score < 0 {
			matePly = -matePly
		}
		fmt.Fprintf(out, "info depth %d nodes %d time %d score mate %d pv %s\n", result.searchDepth(), result.searchNodes(), timeMs, matePly, result.bestMoveUCI())
		return
	}

	fmt.Fprintf(out, "info depth %d nodes %d time %d score cp %d pv %s\n", result.searchDepth(), result.searchNodes(), timeMs, score, result.bestMoveUCI())
}

func sanitizeInfo(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func absScore(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func (s *Server) ensureBestMove(pos *board.Position, result search.Result) search.Result {
	if result.BestMove != (board.Move{}) {
		return result
	}

	root := pos
	if refreshed, err := board.NewPositionFromFEN(pos.FEN()); err == nil {
		root = refreshed
	}

	moves := s.engine.LegalMoves(root)
	if len(moves) > 0 {
		result.BestMove = moves[0]
	}
	return result
}
