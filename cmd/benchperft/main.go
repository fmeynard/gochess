package main

import (
	"chessV2/internal"
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

func main() {
	fen := flag.String("fen", defaultFEN, "FEN to benchmark")
	depth := flag.Int("depth", defaultDepth, "perft depth")
	cpuProfile := flag.String("cpuprofile", ".codex-tmp/bench-perft.cpu.prof", "CPU profile output path")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*cpuProfile), 0o755); err != nil {
		log.Fatalf("create profile directory: %v", err)
	}

	f, err := os.Create(*cpuProfile)
	if err != nil {
		log.Fatalf("create cpu profile: %v", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatalf("start cpu profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	pos, err := internal.NewPositionFromFEN(*fen)
	if err != nil {
		log.Fatalf("parse fen: %v", err)
	}

	engine := internal.NewEngine()
	start := time.Now()
	_, nodes := engine.PerftDivide(pos, *depth)
	elapsed := time.Since(start)

	fmt.Printf("FEN: %s\n", *fen)
	fmt.Printf("Depth: %d\n", *depth)
	fmt.Printf("Nodes: %d\n", nodes)
	fmt.Printf("Elapsed: %s\n", elapsed)
	fmt.Printf("CPU profile: %s\n", *cpuProfile)
}
