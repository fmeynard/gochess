package main

import (
	"bufio"
	"chessV2/internal/match"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type swing struct {
	match.MoveRecord
	BeforeCP int
	AfterCP  int
	DeltaCP  int
	BestMove string
}

type stockfishClient struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	lines    chan string
	readErrs chan error
}

func main() {
	var (
		inputPath     string
		stockfishPath string
		movetimeMS    int
		limit         int
		minSwingCP    int
		onlyCurrent   bool
	)

	flag.StringVar(&inputPath, "input", "", "Path to match JSONL record file")
	flag.StringVar(&stockfishPath, "stockfish", "stockfish", "Path to stockfish binary")
	flag.IntVar(&movetimeMS, "movetime", 250, "Stockfish analysis time per position in milliseconds")
	flag.IntVar(&limit, "limit", 20, "Maximum number of swings to print")
	flag.IntVar(&minSwingCP, "min-swing", 150, "Minimum centipawn swing to include")
	flag.BoolVar(&onlyCurrent, "only-current", true, "Only report swings from current engine moves")
	flag.Parse()

	if inputPath == "" {
		fmt.Fprintln(os.Stderr, "missing required -input")
		os.Exit(2)
	}

	records, err := loadRecords(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client, err := newStockfishClient(stockfishPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer client.close()

	swings := make([]swing, 0, len(records))
	for _, record := range records {
		if onlyCurrent && record.Player != "current" {
			continue
		}

		beforeCP, bestMove, err := client.evaluate(record.FENBefore, time.Duration(movetimeMS)*time.Millisecond)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		afterCP, _, err := client.evaluate(record.FENAfter, time.Duration(movetimeMS)*time.Millisecond)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		delta := beforeCP + afterCP
		if delta < minSwingCP {
			continue
		}

		swings = append(swings, swing{
			MoveRecord: record,
			BeforeCP:   beforeCP,
			AfterCP:    afterCP,
			DeltaCP:    delta,
			BestMove:   bestMove,
		})
	}

	sort.Slice(swings, func(i, j int) bool {
		if swings[i].DeltaCP == swings[j].DeltaCP {
			if swings[i].GameIndex == swings[j].GameIndex {
				return swings[i].Ply < swings[j].Ply
			}
			return swings[i].GameIndex < swings[j].GameIndex
		}
		return swings[i].DeltaCP > swings[j].DeltaCP
	})

	if limit > len(swings) {
		limit = len(swings)
	}

	for i := 0; i < limit; i++ {
		item := swings[i]
		fmt.Printf(
			"#%d game=%d ply=%d delta=%dcp before=%dcp after=%dcp player=%s move=%s sf_best=%s\n",
			i+1,
			item.GameIndex,
			item.Ply,
			item.DeltaCP,
			item.BeforeCP,
			item.AfterCP,
			item.Player,
			item.Move,
			item.BestMove,
		)
		fmt.Printf("  before: %s\n", item.FENBefore)
		fmt.Printf("  after : %s\n", item.FENAfter)
	}
}

func loadRecords(path string) ([]match.MoveRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records := make([]match.MoveRecord, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record match.MoveRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func newStockfishClient(path string) (*stockfishClient, error) {
	cmd := exec.Command(path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	client := &stockfishClient{
		cmd:      cmd,
		stdin:    stdin,
		lines:    make(chan string, 256),
		readErrs: make(chan error, 2),
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go readLines(stdout, client.lines, client.readErrs)
	go drainStream(stderr, client.readErrs)

	if err := client.send("uci"); err != nil {
		return nil, err
	}
	if _, err := client.waitFor("uciok", 5*time.Second); err != nil {
		return nil, err
	}
	if err := client.send("isready"); err != nil {
		return nil, err
	}
	if _, err := client.waitFor("readyok", 5*time.Second); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *stockfishClient) evaluate(fen string, moveTime time.Duration) (int, string, error) {
	if err := c.send("position fen " + fen); err != nil {
		return 0, "", err
	}
	if err := c.send(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
		return 0, "", err
	}

	bestScore := 0
	bestMove := ""
	timer := time.NewTimer(moveTime + 5*time.Second)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return 0, "", c.readFailure("analysis")
			}
			if strings.HasPrefix(line, "info ") {
				if score, ok := parseCentipawnScore(line); ok {
					bestScore = score
				}
				continue
			}
			if strings.HasPrefix(line, "bestmove ") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					bestMove = fields[1]
				}
				return bestScore, bestMove, nil
			}
		case <-timer.C:
			return 0, "", fmt.Errorf("timeout waiting for stockfish bestmove")
		}
	}
}

func (c *stockfishClient) close() error {
	_, _ = fmt.Fprintln(c.stdin, "quit")
	return c.cmd.Wait()
}

func (c *stockfishClient) send(line string) error {
	_, err := fmt.Fprintln(c.stdin, line)
	return err
}

func (c *stockfishClient) waitFor(prefix string, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return "", c.readFailure(fmt.Sprintf("waiting for %q", prefix))
			}
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-timer.C:
			return "", fmt.Errorf("timeout waiting for %q", prefix)
		}
	}
}

func parseCentipawnScore(line string) (int, bool) {
	fields := strings.Fields(line)
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] != "score" || i+2 >= len(fields) {
			continue
		}
		switch fields[i+1] {
		case "cp":
			value, err := strconv.Atoi(fields[i+2])
			if err != nil {
				return 0, false
			}
			return value, true
		case "mate":
			value, err := strconv.Atoi(fields[i+2])
			if err != nil {
				return 0, false
			}
			if value > 0 {
				return 30000, true
			}
			return -30000, true
		}
	}
	return 0, false
}

func (c *stockfishClient) readFailure(context string) error {
	var readErr error
	select {
	case readErr = <-c.readErrs:
	default:
	}
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fmt.Errorf("stockfish reader failed while %s: %w", context, readErr)
	}
	return fmt.Errorf("stockfish exited during %s", context)
}

func readLines(r io.Reader, out chan<- string, errs chan<- error) {
	reader := bufio.NewReader(r)
	defer close(out)

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			out <- strings.TrimSpace(line)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				select {
				case errs <- err:
				default:
				}
			}
			return
		}
	}
}

func drainStream(r io.Reader, errs chan<- error) {
	_, err := io.Copy(io.Discard, r)
	if err != nil && !errors.Is(err, io.EOF) {
		select {
		case errs <- err:
		default:
		}
	}
}
