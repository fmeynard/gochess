package main

import (
	"chessV2/internal/match"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const maxDisplayedGames = 18

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
	fmt.Printf("Markdown: %s", summary.MarkdownRow())
}

func renderSnapshot(snapshot match.Snapshot) {
	var b strings.Builder
	b.WriteString("\033[H\033[2J")
	b.WriteString(fmt.Sprintf("Current: %s\n", snapshot.Current))
	b.WriteString(fmt.Sprintf("Opponent: %s\n", snapshot.Opponent))
	b.WriteString(fmt.Sprintf("Movetime: %s\n\n", snapshot.MoveTime))

	b.WriteString(fmt.Sprintf("Progress  %s %d/%d\n", progressBar(snapshot.CompletedGames, snapshot.TotalGames, 32), snapshot.CompletedGames, snapshot.TotalGames))
	b.WriteString(fmt.Sprintf("Global    %.1f/%d   W/D/L %s\n", snapshot.Score, snapshot.TotalGames, snapshot.Global.Summary()))
	b.WriteString(fmt.Sprintf("As White  W/D/L %s\n", snapshot.AsWhite.Summary()))
	b.WriteString(fmt.Sprintf("As Black  W/D/L %s\n", snapshot.AsBlack.Summary()))
	b.WriteString(fmt.Sprintf("Running   %d\n", snapshot.RunningGames))
	b.WriteString(fmt.Sprintf("Avg NPS   %.0f\n", snapshot.AverageNPS))
	b.WriteString(fmt.Sprintf("Elapsed   %s\n", snapshot.Elapsed.Round(time.Second)))
	b.WriteString(fmt.Sprintf("ETA       %s\n\n", snapshot.EstimatedRemain.Round(time.Second)))

	b.WriteString("Game  Color  Status   Reason               Plies  Time\n")
	b.WriteString("----  -----  -------  -------------------  -----  -------\n")

	start := 0
	if len(snapshot.Games) > maxDisplayedGames {
		start = len(snapshot.Games) - maxDisplayedGames
	}
	for _, game := range snapshot.Games[start:] {
		b.WriteString(fmt.Sprintf(
			"%4d  %-5s  %-7s  %-19s  %5d  %7s\n",
			game.GameIndex,
			colorLabel(game.CurrentAsWhite),
			game.Status,
			truncate(game.Reason, 19),
			game.Plies,
			game.Duration.Round(time.Second),
		))
	}

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
