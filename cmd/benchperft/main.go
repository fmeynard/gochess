package main

import (
	chess "chessV2/internal/chess"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

const (
	defaultFEN   = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
	defaultDepth = 6
)

func runPerft(pos *chess.Position, engine *chess.Engine, depth int) uint64 {
	_, nodes := engine.PerftDivide(pos, depth)
	return nodes
}

func main() {
	fen := flag.String("fen", defaultFEN, "FEN to benchmark")
	depth := flag.Int("depth", defaultDepth, "perft depth")
	cpuProfile := flag.String("cpuprofile", ".codex-tmp/bench-perft.cpu.prof", "CPU profile output path")
	noPerftTricks := flag.Bool("no-perft-tricks", false, "disable perft-only tricks such as bulk counting and transposition table use")
	mode := flag.String("mode", "hot", "benchmark mode: hot excludes FEN/engine setup from the timed section, cold includes it")
	warmup := flag.Int("warmup", 1, "number of warmup perft runs before timing in hot mode")
	flag.Parse()

	if *mode != "hot" && *mode != "cold" {
		log.Fatalf("invalid benchmark mode %q", *mode)
	}
	if *mode == "cold" && *warmup != 0 {
		log.Fatalf("warmup is only supported in hot mode")
	}

	var (
		profileFile *os.File
		err         error
	)
	if *cpuProfile != "" {
		if err := os.MkdirAll(filepath.Dir(*cpuProfile), 0o755); err != nil {
			log.Fatalf("create profile directory: %v", err)
		}
		profileFile, err = os.Create(*cpuProfile)
		if err != nil {
			log.Fatalf("create cpu profile: %v", err)
		}
		defer profileFile.Close()
	}

	nodes := uint64(0)
	start := time.Now()

	switch *mode {
	case "hot":
		pos, err := chess.NewPositionFromFEN(*fen)
		if err != nil {
			log.Fatalf("parse fen: %v", err)
		}

		engine := chess.NewEngine()
		engine.SetPerftTricks(!*noPerftTricks)

		for i := 0; i < *warmup; i++ {
			warmPos, err := chess.NewPositionFromFEN(*fen)
			if err != nil {
				log.Fatalf("parse fen for warmup: %v", err)
			}
			runPerft(warmPos, engine, *depth)
		}

		if profileFile != nil {
			if err := pprof.StartCPUProfile(profileFile); err != nil {
				log.Fatalf("start cpu profile: %v", err)
			}
		}
		start = time.Now()
		nodes = runPerft(pos, engine, *depth)
		if profileFile != nil {
			pprof.StopCPUProfile()
		}
	case "cold":
		if profileFile != nil {
			if err := pprof.StartCPUProfile(profileFile); err != nil {
				log.Fatalf("start cpu profile: %v", err)
			}
		}
		start = time.Now()
		pos, err := chess.NewPositionFromFEN(*fen)
		if err != nil {
			log.Fatalf("parse fen: %v", err)
		}

		engine := chess.NewEngine()
		engine.SetPerftTricks(!*noPerftTricks)
		nodes = runPerft(pos, engine, *depth)
		if profileFile != nil {
			pprof.StopCPUProfile()
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("FEN: %s\n", *fen)
	fmt.Printf("Depth: %d\n", *depth)
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Printf("Warmup: %d\n", *warmup)
	fmt.Printf("Perft tricks: %t\n", !*noPerftTricks)
	fmt.Printf("Nodes: %d\n", nodes)
	fmt.Printf("Elapsed: %s\n", elapsed)
	if *cpuProfile != "" {
		fmt.Printf("CPU profile: %s\n", *cpuProfile)
	}
}
