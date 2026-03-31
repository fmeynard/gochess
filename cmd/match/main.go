package main

import (
	"chessV2/internal/match"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var opponentTag string
	var games int
	var moveTimeMs int
	var notes string

	flag.StringVar(&opponentTag, "opponent-tag", "", "Git tag to build and use as the opponent")
	flag.IntVar(&games, "games", 2, "Number of games to play")
	flag.IntVar(&moveTimeMs, "movetime", 5000, "Per-move time budget in milliseconds")
	flag.StringVar(&notes, "notes", "", "Optional notes to include in the printed markdown row")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	summary, err := match.RunMatch(match.Config{
		RepoRoot:    repoRoot,
		OpponentTag: opponentTag,
		Games:       games,
		MoveTime:    time.Duration(moveTimeMs) * time.Millisecond,
		Notes:       notes,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current: %s\n", summary.Current)
	fmt.Printf("Opponent: %s\n", summary.Opponent)
	fmt.Printf("Movetime: %s\n", summary.MoveTime)
	fmt.Printf("Games: %d\n", summary.Games)
	fmt.Printf("Score: %s\n", summary.ScoreSummary())
	fmt.Printf("W/D/L: %s\n", summary.WDLSummary())
	fmt.Printf("Markdown: %s", summary.MarkdownRow())
}
