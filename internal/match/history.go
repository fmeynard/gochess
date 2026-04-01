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
	AsWhite       Record
	AsBlack       Record
	Reasons       map[string]int
	IllegalMoves  []IllegalMoveDiagnostic
	Notes         string
}

type IllegalMoveDiagnostic struct {
	GameIndex      int
	CurrentAsWhite bool
	Offender       string
	BestMove       string
	FEN            string
	LegalMoves     []string
	LastPosition   string
	LastGo         string
	BestMoveRaw    string
	RecentLines    []string
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

func (s Summary) ReasonSummary() string {
	if len(s.Reasons) == 0 {
		return "-"
	}

	order := []string{
		"checkmate",
		"draw by repetition",
		"stalemate",
		"max plies",
		"illegal move",
		"search error",
	}

	parts := make([]string, 0, len(s.Reasons))
	seen := make(map[string]bool, len(order))
	for _, reason := range order {
		count := s.Reasons[reason]
		if count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%d", reason, count))
		seen[reason] = true
	}

	for reason, count := range s.Reasons {
		if seen[reason] || count == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%d", reason, count))
	}

	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

func escapeMarkdownCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
