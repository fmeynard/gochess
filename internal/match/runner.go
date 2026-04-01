package match

import (
	board "chessV2/internal/board"
	"chessV2/internal/engine"
	"chessV2/internal/movegen"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const DefaultMoveTime = 5 * time.Second
const defaultMaxPlies = 300

type Config struct {
	RepoRoot      string
	OpponentTag   string
	Games         int
	Parallelism   int
	MoveTime      time.Duration
	Notes         string
	CurrentLabel  string
	CurrentBinary string
	Progress      func(Snapshot)
}

type binarySpec struct {
	Label string
	Path  string
}

type GameRecord struct {
	GameIndex      int
	CurrentAsWhite bool
	Status         string
	Reason         string
	Plies          int
	Duration       time.Duration
	Diagnostic     *IllegalMoveDiagnostic
}

type Snapshot struct {
	Current         string
	Opponent        string
	MoveTime        time.Duration
	TotalGames      int
	CompletedGames  int
	RunningGames    int
	Elapsed         time.Duration
	EstimatedRemain time.Duration
	Score           float64
	AverageNPS      float64
	Global          Record
	AsWhite         Record
	AsBlack         Record
	Games           []GameRecord
}

type gameResult struct {
	gameIndex      int
	currentAsWhite bool
	outcome        int
	reason         string
	plies          int
	duration       time.Duration
	nodes          uint64
	searchTime     time.Duration
	diagnostic     *IllegalMoveDiagnostic
}

type matchState struct {
	mu      sync.Mutex
	start   time.Time
	cfg     Config
	summary Summary
	asWhite Record
	asBlack Record
	games   []GameRecord
	started []time.Time
	running int
	nodes   uint64
	time    time.Duration
}

func RunMatch(cfg Config) (Summary, error) {
	if cfg.Games <= 0 {
		return Summary{}, fmt.Errorf("games must be > 0")
	}
	if cfg.MoveTime <= 0 {
		cfg.MoveTime = DefaultMoveTime
	}
	if cfg.Parallelism <= 0 {
		cfg.Parallelism = 1
	}
	if cfg.Parallelism > cfg.Games {
		cfg.Parallelism = cfg.Games
	}

	currentLabel, err := cfg.currentRevision()
	if err != nil {
		return Summary{}, err
	}
	if cfg.CurrentLabel == "" {
		cfg.CurrentLabel = currentLabel
	}

	currentBinary, err := cfg.prepareCurrentBinary()
	if err != nil {
		return Summary{}, err
	}

	opponent, err := cfg.prepareOpponentBinary(currentBinary)
	if err != nil {
		return Summary{}, err
	}

	state := newMatchState(cfg, opponent.Label)
	state.emit()

	stopTicker := make(chan struct{})
	var tickerWG sync.WaitGroup
	if cfg.Progress != nil {
		tickerWG.Add(1)
		go func() {
			defer tickerWG.Done()
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					state.emit()
				case <-stopTicker:
					return
				}
			}
		}()
	}
	defer func() {
		close(stopTicker)
		tickerWG.Wait()
	}()

	jobs := make(chan int)
	results := make(chan gameResult, cfg.Games)
	errs := make(chan error, 1)

	var workers sync.WaitGroup
	for workerID := 0; workerID < cfg.Parallelism; workerID++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := runWorker(currentBinary.Path, opponent.Path, cfg.MoveTime, jobs, results, state); err != nil {
				select {
				case errs <- err:
				default:
				}
			}
		}()
	}

	go func() {
		for gameIndex := 0; gameIndex < cfg.Games; gameIndex++ {
			jobs <- gameIndex
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()

	completed := 0
	for completed < cfg.Games {
		select {
		case err := <-errs:
			if err != nil {
				return Summary{}, err
			}
		case result, ok := <-results:
			if !ok {
				completed = cfg.Games
				continue
			}
			state.finish(result)
			completed++
		}
	}

	return state.summary, nil
}

func runWorker(currentPath, opponentPath string, moveTime time.Duration, jobs <-chan int, results chan<- gameResult, state *matchState) error {
	currentClient, err := NewUCIClient(currentPath)
	if err != nil {
		return err
	}
	defer currentClient.Close()

	opponentClient, err := NewUCIClient(opponentPath)
	if err != nil {
		return err
	}
	defer opponentClient.Close()

	for gameIndex := range jobs {
		if err := currentClient.NewGame(); err != nil {
			return err
		}
		if err := opponentClient.NewGame(); err != nil {
			return err
		}

		currentAsWhite := gameIndex%2 == 0
		state.startGame(gameIndex, currentAsWhite)

		start := time.Now()
		outcome, plies, reason, nodes, searchTime, diagnostic, err := playSingleGame(
			currentClient,
			opponentClient,
			currentAsWhite,
			moveTime,
			func(ply int) {
				state.updatePlies(gameIndex, ply)
			},
		)
		if err != nil {
			return err
		}
		if diagnostic != nil {
			diagnostic.GameIndex = gameIndex + 1
		}

		results <- gameResult{
			gameIndex:      gameIndex,
			currentAsWhite: currentAsWhite,
			outcome:        outcome,
			reason:         reason,
			plies:          plies,
			duration:       time.Since(start),
			nodes:          nodes,
			searchTime:     searchTime,
			diagnostic:     diagnostic,
		}
	}

	return nil
}

func newMatchState(cfg Config, opponentLabel string) *matchState {
	games := make([]GameRecord, cfg.Games)
	for i := range games {
		games[i] = GameRecord{
			GameIndex:      i + 1,
			CurrentAsWhite: i%2 == 0,
			Status:         "pending",
			Reason:         "-",
		}
	}

	return &matchState{
		start: time.Now(),
		cfg:   cfg,
		summary: Summary{
			Date:     time.Now().UTC(),
			Current:  cfg.CurrentLabel,
			Opponent: opponentLabel,
			MoveTime: cfg.MoveTime,
			Games:    cfg.Games,
			Reasons:  make(map[string]int),
			Notes:    cfg.Notes,
		},
		games:   games,
		started: make([]time.Time, cfg.Games),
	}
}

func (s *matchState) startGame(gameIndex int, currentAsWhite bool) {
	s.mu.Lock()
	s.games[gameIndex].Status = "running"
	s.games[gameIndex].Reason = "-"
	s.games[gameIndex].Plies = 0
	s.games[gameIndex].Duration = 0
	s.games[gameIndex].CurrentAsWhite = currentAsWhite
	s.started[gameIndex] = time.Now()
	s.running++
	s.mu.Unlock()
	s.emit()
}

func (s *matchState) finish(result gameResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := &s.games[result.gameIndex]
	record.Status = outcomeLabel(result.outcome)
	record.Reason = result.reason
	record.Plies = result.plies
	record.Duration = result.duration
	record.Diagnostic = result.diagnostic
	s.started[result.gameIndex] = time.Time{}

	s.running--
	s.nodes += result.nodes
	s.time += result.searchTime
	s.summary.Reasons[result.reason]++
	if result.diagnostic != nil {
		s.summary.IllegalMoves = append(s.summary.IllegalMoves, *result.diagnostic)
	}
	switch result.outcome {
	case 1:
		s.summary.CurrentWins++
		if result.currentAsWhite {
			s.asWhite.Wins++
			s.summary.AsWhite = s.asWhite
		} else {
			s.asBlack.Wins++
			s.summary.AsBlack = s.asBlack
		}
	case -1:
		s.summary.CurrentLosses++
		if result.currentAsWhite {
			s.asWhite.Losses++
			s.summary.AsWhite = s.asWhite
		} else {
			s.asBlack.Losses++
			s.summary.AsBlack = s.asBlack
		}
	default:
		s.summary.Draws++
		if result.currentAsWhite {
			s.asWhite.Draws++
			s.summary.AsWhite = s.asWhite
		} else {
			s.asBlack.Draws++
			s.summary.AsBlack = s.asBlack
		}
	}

	s.emitLocked()
}

func (s *matchState) emit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emitLocked()
}

func (s *matchState) emitLocked() {
	if s.cfg.Progress == nil {
		return
	}

	completed := s.summary.CurrentWins + s.summary.CurrentLosses + s.summary.Draws
	elapsed := time.Since(s.start)
	remainingGames := s.cfg.Games - completed
	var eta time.Duration
	if completed > 0 {
		eta = (elapsed / time.Duration(completed)) * time.Duration(remainingGames)
	}

	gamesCopy := make([]GameRecord, len(s.games))
	copy(gamesCopy, s.games)
	now := time.Now()
	for i := range gamesCopy {
		if gamesCopy[i].Status == "running" && !s.started[i].IsZero() {
			gamesCopy[i].Duration = now.Sub(s.started[i])
		}
	}

	s.cfg.Progress(Snapshot{
		Current:         s.summary.Current,
		Opponent:        s.summary.Opponent,
		MoveTime:        s.summary.MoveTime,
		TotalGames:      s.cfg.Games,
		CompletedGames:  completed,
		RunningGames:    s.running,
		Elapsed:         elapsed,
		EstimatedRemain: eta,
		Score:           s.summary.Score(),
		AverageNPS:      averageNPS(s.nodes, s.time),
		Global: Record{
			Wins:   s.summary.CurrentWins,
			Draws:  s.summary.Draws,
			Losses: s.summary.CurrentLosses,
		},
		AsWhite: s.asWhite,
		AsBlack: s.asBlack,
		Games:   gamesCopy,
	})
}

func playSingleGame(currentClient, opponentClient *UCIClient, currentIsWhite bool, moveTime time.Duration, onPly func(int)) (int, int, string, uint64, time.Duration, *IllegalMoveDiagnostic, error) {
	referee := engine.NewEngine()
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return 0, 0, "", 0, 0, nil, err
	}

	moves := make([]string, 0, defaultMaxPlies)
	repetitionCount := map[uint64]int{pos.ZobristKey(): 1}
	var totalNodes uint64
	var totalSearchTime time.Duration

	for ply := 0; ply < defaultMaxPlies; ply++ {
		if repetitionCount[pos.ZobristKey()] >= 3 {
			return 0, ply, "draw by repetition", totalNodes, totalSearchTime, nil, nil
		}

		legalMoves := referee.LegalMoves(pos)
		if len(legalMoves) == 0 {
			if movegen.IsKingInCheck(pos, pos.ActiveColor()) {
				currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
				if currentToMove {
					return -1, ply, "checkmate", totalNodes, totalSearchTime, nil, nil
				}
				return 1, ply, "checkmate", totalNodes, totalSearchTime, nil, nil
			}
			return 0, ply, "stalemate", totalNodes, totalSearchTime, nil, nil
		}

		client := selectClient(currentClient, opponentClient, pos.ActiveColor(), currentIsWhite)
		bestMove, stats, err := client.BestMove(moves, moveTime)
		if err != nil {
			currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
			if currentToMove {
				return -1, ply, "search error", totalNodes, totalSearchTime, nil, nil
			}
			return 1, ply, "search error", totalNodes, totalSearchTime, nil, nil
		}
		totalNodes += stats.Nodes
		totalSearchTime += stats.Time

		if err := referee.ApplyUCIMove(pos, bestMove); err != nil {
			currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
			diagnostic := &IllegalMoveDiagnostic{
				GameIndex:      0,
				CurrentAsWhite: currentIsWhite,
				Offender:       illegalMoveOffender(currentToMove),
				BestMove:       bestMove,
				FEN:            pos.FEN(),
				LegalMoves:     boardMovesToUCIs(legalMoves),
			}
			if currentToMove {
				return -1, ply, "illegal move", totalNodes, totalSearchTime, diagnostic, nil
			}
			return 1, ply, "illegal move", totalNodes, totalSearchTime, diagnostic, nil
		}

		moves = append(moves, bestMove)
		repetitionCount[pos.ZobristKey()]++
		if onPly != nil {
			onPly(ply + 1)
		}
	}

	return 0, defaultMaxPlies, "max plies", totalNodes, totalSearchTime, nil, nil
}

func (s *matchState) updatePlies(gameIndex int, plies int) {
	s.mu.Lock()
	s.games[gameIndex].Plies = plies
	s.mu.Unlock()
}

func selectClient(currentClient, opponentClient *UCIClient, activeColor int8, currentIsWhite bool) *UCIClient {
	currentToMove := (activeColor == board.White) == currentIsWhite
	if currentToMove {
		return currentClient
	}
	return opponentClient
}

func (c Config) prepareCurrentBinary() (binarySpec, error) {
	if c.CurrentBinary != "" {
		return binarySpec{Label: c.CurrentLabel, Path: c.CurrentBinary}, nil
	}

	outputPath := filepath.Join(c.RepoRoot, ".codex-tmp", "match", "current", "gochess-uci")
	if err := buildUCIBinary(c.RepoRoot, c.RepoRoot, outputPath); err != nil {
		return binarySpec{}, err
	}

	return binarySpec{Label: c.CurrentLabel, Path: outputPath}, nil
}

func (c Config) prepareOpponentBinary(current binarySpec) (binarySpec, error) {
	if c.OpponentTag == "" {
		return binarySpec{
			Label: fmt.Sprintf("self@%s", current.Label),
			Path:  current.Path,
		}, nil
	}

	safeTag := sanitizeForPath(c.OpponentTag)
	worktreePath := filepath.Join(c.RepoRoot, ".codex-tmp", "match", "worktrees", safeTag)
	outputPath := filepath.Join(c.RepoRoot, ".codex-tmp", "match", "opponents", safeTag, "gochess-uci")

	if err := os.RemoveAll(worktreePath); err != nil {
		return binarySpec{}, err
	}

	if err := runCommand(c.RepoRoot, "git", "worktree", "add", "--detach", worktreePath, c.OpponentTag); err != nil {
		return binarySpec{}, err
	}
	defer runCommand(c.RepoRoot, "git", "worktree", "remove", "--force", worktreePath)

	if err := buildUCIBinary(c.RepoRoot, worktreePath, outputPath); err != nil {
		return binarySpec{}, fmt.Errorf("failed to build opponent tag %s: %w", c.OpponentTag, err)
	}

	opponentRev, err := revParseShort(c.RepoRoot, c.OpponentTag)
	if err != nil {
		return binarySpec{}, err
	}

	return binarySpec{
		Label: fmt.Sprintf("tag %s (%s)", c.OpponentTag, opponentRev),
		Path:  outputPath,
	}, nil
}

func (c Config) currentRevision() (string, error) {
	return revParseShort(c.RepoRoot, "HEAD")
}

func buildUCIBinary(repoRoot, workdir, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/uci")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(repoRoot, ".codex-tmp", "go-build-cache"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runCommand(workdir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func revParseShort(workdir string, rev string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", rev)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizeForPath(s string) string {
	replacer := strings.NewReplacer("/", "-", " ", "-", ":", "-", "@", "-", "\\", "-")
	return replacer.Replace(s)
}

func outcomeLabel(outcome int) string {
	switch outcome {
	case 1:
		return "win"
	case -1:
		return "loss"
	default:
		return "draw"
	}
}

func illegalMoveOffender(currentToMove bool) string {
	if currentToMove {
		return "current"
	}
	return "opponent"
}

func boardMovesToUCIs(moves []board.Move) []string {
	ucis := make([]string, 0, len(moves))
	for _, move := range moves {
		ucis = append(ucis, move.UCI())
	}
	return ucis
}

func averageNPS(nodes uint64, searchTime time.Duration) float64 {
	if searchTime <= 0 {
		return 0
	}
	return float64(nodes) / searchTime.Seconds()
}
