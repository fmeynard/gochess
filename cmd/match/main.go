package main

import (
	"chessV2/internal/match"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const maxDisplayedGames = 10

func main() {
	var opponentTag string
	var games int
	var parallel int
	var moveTimeMs int
	var notes string
	var plain bool

	flag.StringVar(&opponentTag, "opponent-tag", "", "Git tag to build and use as the opponent")
	flag.IntVar(&games, "games", 2, "Number of games to play")
	flag.IntVar(&parallel, "parallel", 1, "Number of games to run concurrently")
	flag.IntVar(&moveTimeMs, "movetime", 5000, "Per-move time budget in milliseconds")
	flag.StringVar(&notes, "notes", "", "Optional notes to include in the printed markdown row")
	flag.BoolVar(&plain, "plain", false, "Use plain line-based progress instead of the live terminal dashboard")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	progress := func(snapshot match.Snapshot) {
		if plain {
			fmt.Printf(
				"[%3d/%3d] running=%d score=%.1f/%d global=%s white=%s black=%s nps=%.0f elapsed=%s eta=%s\n",
				snapshot.CompletedGames,
				snapshot.TotalGames,
				snapshot.RunningGames,
				snapshot.Score,
				snapshot.TotalGames,
				snapshot.Global.Summary(),
				snapshot.AsWhite.Summary(),
				snapshot.AsBlack.Summary(),
				snapshot.AverageNPS,
				snapshot.Elapsed.Round(time.Second),
				snapshot.EstimatedRemain.Round(time.Second),
			)
			return
		}
		renderSnapshot(snapshot)
	}

	summary, err := match.RunMatch(match.Config{
		RepoRoot:    repoRoot,
		OpponentTag: opponentTag,
		Games:       games,
		Parallelism: parallel,
		MoveTime:    time.Duration(moveTimeMs) * time.Millisecond,
		Notes:       notes,
		Progress:    progress,
	})
	if err != nil {
		panic(err)
	}

	if !plain {
		fmt.Print("\033[H\033[2J")
	}

	fmt.Printf("Current: %s\n", summary.Current)
	fmt.Printf("Opponent: %s\n", summary.Opponent)
	fmt.Printf("Movetime: %s\n", summary.MoveTime)
	fmt.Printf("Games: %d\n", summary.Games)
	fmt.Printf("Score: %s\n", summary.ScoreSummary())
	fmt.Printf("W/D/L: %s\n", summary.WDLSummary())
	fmt.Printf("As White W/D/L: %s\n", summary.AsWhite.Summary())
	fmt.Printf("As Black W/D/L: %s\n", summary.AsBlack.Summary())
	fmt.Printf("Reasons: %s\n", summary.ReasonSummary())
	fmt.Printf("Markdown: %s", summary.MarkdownRow())
}

func renderSnapshot(snapshot match.Snapshot) {
	var b strings.Builder
	b.WriteString("\033[H\033[2J")
	b.WriteString(box(
		"Match",
		[]string{
			fmt.Sprintf("Current   %s", snapshot.Current),
			fmt.Sprintf("Opponent  %s", snapshot.Opponent),
			fmt.Sprintf("Movetime  %s", snapshot.MoveTime),
		},
	))
	b.WriteString("\n")
	b.WriteString(box(
		"Stats",
		[]string{
			fmt.Sprintf("Progress  %s %d/%d", progressBar(snapshot.CompletedGames, snapshot.TotalGames, 28), snapshot.CompletedGames, snapshot.TotalGames),
			fmt.Sprintf("Global    %.1f/%d   W/D/L %s", snapshot.Score, snapshot.TotalGames, snapshot.Global.Summary()),
			fmt.Sprintf("As White  W/D/L %s", snapshot.AsWhite.Summary()),
			fmt.Sprintf("As Black  W/D/L %s", snapshot.AsBlack.Summary()),
			fmt.Sprintf("Running   %d", snapshot.RunningGames),
			fmt.Sprintf("Avg NPS   %.0f", snapshot.AverageNPS),
			fmt.Sprintf("Elapsed   %s", snapshot.Elapsed.Round(time.Second)),
			fmt.Sprintf("ETA       %s", snapshot.EstimatedRemain.Round(time.Second)),
		},
	))
	b.WriteString("\n")
	b.WriteString("Games\n")
	b.WriteString("┌──────┬───────┬─────────┬─────────────────────┬───────┬────────┐\n")
	b.WriteString("│ Game │ Color │ Status  │ Reason              │ Plies │ Time   │\n")
	b.WriteString("├──────┼───────┼─────────┼─────────────────────┼───────┼────────┤\n")

	for _, game := range displayedGames(snapshot.Games) {
		b.WriteString(fmt.Sprintf(
			"│ %4d │ %-5s │ %-7s │ %-19s │ %5d │ %6s │\n",
			game.GameIndex,
			colorLabel(game.CurrentAsWhite),
			game.Status,
			truncate(game.Reason, 19),
			game.Plies,
			game.Duration.Round(time.Second),
		))
	}
	b.WriteString("└──────┴───────┴─────────┴─────────────────────┴───────┴────────┘\n")

	fmt.Print(b.String())
}

func progressBar(done, total, width int) string {
	if total <= 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}
	filled := width * done / total
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func colorLabel(currentAsWhite bool) string {
	if currentAsWhite {
		return "white"
	}
	return "black"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func displayedGames(games []match.GameRecord) []match.GameRecord {
	selected := make([]match.GameRecord, 0, maxDisplayedGames)

	for _, game := range games {
		if game.Status == "running" {
			selected = append(selected, game)
		}
	}

	if len(selected) < maxDisplayedGames {
		for _, game := range games {
			if game.Status == "pending" {
				selected = append(selected, game)
				if len(selected) == maxDisplayedGames {
					return selected
				}
			}
		}
	}

	if len(selected) == 0 {
		start := 0
		if len(games) > maxDisplayedGames {
			start = len(games) - maxDisplayedGames
		}
		return games[start:]
	}

	if len(selected) > maxDisplayedGames {
		return selected[:maxDisplayedGames]
	}

	return selected
}

func box(title string, lines []string) string {
	width := len(title)
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	var b strings.Builder
	b.WriteString("┌─ ")
	b.WriteString(title)
	b.WriteString(" ")
	b.WriteString(strings.Repeat("─", width-len(title)))
	b.WriteString("─┐\n")
	for _, line := range lines {
		b.WriteString("│ ")
		b.WriteString(line)
		if pad := width - len(line); pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString(" │\n")
	}
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", width+2))
	b.WriteString("┘\n")
	return b.String()
}
