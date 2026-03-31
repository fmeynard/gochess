package match

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type UCIClient struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines chan string
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

func (c *UCIClient) BestMove(moves []string, moveTime time.Duration) (string, error) {
	position := "position startpos"
	if len(moves) > 0 {
		position += " moves " + strings.Join(moves, " ")
	}
	if err := c.send(position); err != nil {
		return "", err
	}
	if err := c.send(fmt.Sprintf("go movetime %d", moveTime.Milliseconds())); err != nil {
		return "", err
	}

	line, err := c.waitFor("bestmove ", moveTime+5*time.Second)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", fmt.Errorf("invalid bestmove response: %q", line)
	}
	return fields[1], nil
}

func (c *UCIClient) Close() error {
	_ = c.send("quit")
	return c.cmd.Wait()
}

func (c *UCIClient) send(line string) error {
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
			if strings.HasPrefix(line, prefix) {
				return line, nil
			}
		case <-timer.C:
			return "", fmt.Errorf("timeout waiting for %q", prefix)
		}
	}
}

func scanLines(r io.Reader, out chan<- string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out <- strings.TrimSpace(scanner.Text())
	}
}
