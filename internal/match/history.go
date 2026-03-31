package match

import (
	"fmt"
	"strings"
	"time"
)

type Summary struct {
	Date          time.Time
	Current       string
	Opponent      string
	MoveTime      time.Duration
	Games         int
	CurrentWins   int
	Draws         int
	CurrentLosses int
	Notes         string
}

type Record struct {
	Wins   int
	Draws  int
	Losses int
}

func (r Record) Summary() string {
	return fmt.Sprintf("%d/%d/%d", r.Wins, r.Draws, r.Losses)
}

func (s Summary) Score() float64 {
	return float64(s.CurrentWins) + 0.5*float64(s.Draws)
}

func (s Summary) WDLSummary() string {
	return fmt.Sprintf("%d/%d/%d", s.CurrentWins, s.Draws, s.CurrentLosses)
}

func (s Summary) ScoreSummary() string {
	return fmt.Sprintf("%.1f/%d", s.Score(), s.Games)
}

func (s Summary) MarkdownRow() string {
	return fmt.Sprintf(
		"| %s | %s | %s | %s | %d | %s | %s | %s |\n",
		s.Date.UTC().Format(time.RFC3339),
		escapeMarkdownCell(s.Current),
		escapeMarkdownCell(s.Opponent),
		s.MoveTime.Round(time.Millisecond),
		s.Games,
		s.ScoreSummary(),
		s.WDLSummary(),
		escapeMarkdownCell(s.Notes),
	)
}

func escapeMarkdownCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
