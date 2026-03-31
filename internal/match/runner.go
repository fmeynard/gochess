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
	"time"
)

const DefaultMoveTime = 5 * time.Second
const defaultMaxPlies = 300

type Config struct {
	RepoRoot      string
	OpponentTag   string
	Games         int
	MoveTime      time.Duration
	Notes         string
	CurrentLabel  string
	CurrentBinary string
}

type binarySpec struct {
	Label string
	Path  string
}

func RunMatch(cfg Config) (Summary, error) {
	if cfg.Games <= 0 {
		return Summary{}, fmt.Errorf("games must be > 0")
	}
	if cfg.MoveTime <= 0 {
		cfg.MoveTime = DefaultMoveTime
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

	currentClient, err := NewUCIClient(currentBinary.Path)
	if err != nil {
		return Summary{}, err
	}
	defer currentClient.Close()

	opponentClient, err := NewUCIClient(opponent.Path)
	if err != nil {
		return Summary{}, err
	}
	defer opponentClient.Close()

	summary := Summary{
		Date:     time.Now().UTC(),
		Current:  cfg.CurrentLabel,
		Opponent: opponent.Label,
		MoveTime: cfg.MoveTime,
		Games:    cfg.Games,
		Notes:    cfg.Notes,
	}

	for game := 0; game < cfg.Games; game++ {
		if err := currentClient.NewGame(); err != nil {
			return Summary{}, err
		}
		if err := opponentClient.NewGame(); err != nil {
			return Summary{}, err
		}

		outcome, err := playSingleGame(currentClient, opponentClient, game%2 == 0, cfg.MoveTime)
		if err != nil {
			return Summary{}, err
		}

		switch outcome {
		case 1:
			summary.CurrentWins++
		case -1:
			summary.CurrentLosses++
		default:
			summary.Draws++
		}
	}

	return summary, nil
}

func playSingleGame(currentClient, opponentClient *UCIClient, currentIsWhite bool, moveTime time.Duration) (int, error) {
	referee := engine.NewEngine()
	pos, err := board.NewPositionFromFEN(board.FenStartPos)
	if err != nil {
		return 0, err
	}

	moves := make([]string, 0, defaultMaxPlies)
	repetitionCount := map[uint64]int{pos.ZobristKey(): 1}

	for ply := 0; ply < defaultMaxPlies; ply++ {
		if repetitionCount[pos.ZobristKey()] >= 3 {
			return 0, nil
		}

		legalMoves := referee.LegalMoves(pos)
		if len(legalMoves) == 0 {
			if movegen.IsKingInCheck(pos, pos.ActiveColor()) {
				currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
				if currentToMove {
					return -1, nil
				}
				return 1, nil
			}
			return 0, nil
		}

		client := selectClient(currentClient, opponentClient, pos.ActiveColor(), currentIsWhite)
		bestMove, err := client.BestMove(moves, moveTime)
		if err != nil {
			currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
			if currentToMove {
				return -1, nil
			}
			return 1, nil
		}

		if err := referee.ApplyUCIMove(pos, bestMove); err != nil {
			currentToMove := (pos.ActiveColor() == board.White) == currentIsWhite
			if currentToMove {
				return -1, nil
			}
			return 1, nil
		}

		moves = append(moves, bestMove)
		repetitionCount[pos.ZobristKey()]++
	}

	return 0, nil
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
