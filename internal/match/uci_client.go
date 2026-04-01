package match

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UCIClient struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines chan string

	mu           sync.Mutex
	lastPosition string
	lastGo       string
	bestMoveRaw  string
	recentLines  []string
}

type SearchStats struct {
	Nodes uint64
	Time  time.Duration
}

type ClientDebugSnapshot struct {
	LastPosition string
	LastGo       string
	BestMoveRaw  string
	RecentLines  []string
}

func NewUCIClient(binaryPath string) (*UCIClient, error) {
	cmd := exec.Command(binaryPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	client := &UCIClient{
		cmd:   cmd,
		stdin: stdin,
		lines: make(chan string, 256),
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go scanLines(stdout, client.lines)
	go scanLines(stderr, client.lines)

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

func (c *UCIClient) NewGame() error {
	if err := c.send("ucinewgame"); err != nil {
		return err
	}
	if err := c.send("isready"); err != nil {
		return err
	}
	_, err := c.waitFor("readyok", 5*time.Second)
	return err
}

func (c *UCIClient) BestMove(moves []string, moveTime time.Duration) (string, SearchStats, error) {
	if err := c.ready(); err != nil {
		return "", SearchStats{}, err
	}

	position := "position startpos"
	if len(moves) > 0 {
		position += " moves " + strings.Join(moves, " ")
	}
	if err := c.send(position); err != nil {
		return "", SearchStats{}, err
	}
	if err := c.send(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
		return "", SearchStats{}, err
	}

	line, stats, err := c.waitForBestMove(moveTime + 5*time.Second)
	if err != nil {
		return "", SearchStats{}, err
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", SearchStats{}, fmt.Errorf("invalid bestmove response: %q", line)
	}
	return fields[1], stats, nil
}

func (c *UCIClient) DebugSnapshot() ClientDebugSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	lines := make([]string, len(c.recentLines))
	copy(lines, c.recentLines)
	return ClientDebugSnapshot{
		LastPosition: c.lastPosition,
		LastGo:       c.lastGo,
		BestMoveRaw:  c.bestMoveRaw,
		RecentLines:  lines,
	}
}

func (c *UCIClient) Close() error {
	_ = c.send("quit")
	return c.cmd.Wait()
}

func (c *UCIClient) ready() error {
	if err := c.send("isready"); err != nil {
		return err
	}
	_, err := c.waitFor("readyok", 5*time.Second)
	return err
}

func (c *UCIClient) send(line string) error {
	c.recordSend(line)
	_, err := fmt.Fprintln(c.stdin, line)
	return err
}

func (c *UCIClient) waitFor(prefix string, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return "", fmt.Errorf("uci process ended before %q", prefix)
			}
			c.recordLine(line)
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-timer.C:
			return "", fmt.Errorf("timeout waiting for %q", prefix)
		}
	}
}

func (c *UCIClient) waitForBestMove(timeout time.Duration) (string, SearchStats, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var stats SearchStats
	for {
		select {
		case line, ok := <-c.lines:
			if !ok {
				return "", SearchStats{}, fmt.Errorf("uci process ended before %q", "bestmove ")
			}
			c.recordLine(line)
			if strings.HasPrefix(line, "info ") {
				stats = parseInfoStats(line, stats)
				continue
			}
			if strings.HasPrefix(line, "bestmove ") {
				return line, stats, nil
			}
		case <-timer.C:
			return "", SearchStats{}, fmt.Errorf("timeout waiting for %q", "bestmove ")
		}
	}
}

func parseInfoStats(line string, stats SearchStats) SearchStats {
	fields := strings.Fields(line)
	for i := 0; i < len(fields)-1; i++ {
		switch fields[i] {
		case "nodes":
			if value, err := strconv.ParseUint(fields[i+1], 10, 64); err == nil {
				stats.Nodes = value
			}
			i++
		case "time":
			if value, err := strconv.ParseInt(fields[i+1], 10, 64); err == nil {
				stats.Time = time.Duration(value) * time.Millisecond
			}
			i++
		}
	}
	return stats
}

func scanLines(r io.Reader, out chan<- string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out <- strings.TrimSpace(scanner.Text())
	}
}

func (c *UCIClient) recordSend(line string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case strings.HasPrefix(line, "position "):
		c.lastPosition = line
	case strings.HasPrefix(line, "go "):
		c.lastGo = line
		c.bestMoveRaw = ""
	}
}

func (c *UCIClient) recordLine(line string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strings.HasPrefix(line, "bestmove ") {
		c.bestMoveRaw = line
	}
	c.recentLines = append(c.recentLines, line)
	if len(c.recentLines) > 12 {
		c.recentLines = c.recentLines[len(c.recentLines)-12:]
	}
}
