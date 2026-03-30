# Benchmark History

This file tracks repeatable performance measurements for the engine and the optimizations introduced between versions.

## Methodology

Baseline profiling and timing were recorded on March 31, 2026 from the local repository using:

- Go version: `go1.26.1-X:nodwarf5 linux/amd64`
- Machine/OS: `Linux patafixd 6.19.10-arch1-1 x86_64`
- Target: perft divide on a complex regression position
- Position label: `Perft position 3`
- FEN: `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1`
- Depth: `6`

Command used to generate the baseline timing and CPU profile:

```bash
./.codex-tmp/perft-profiler \
  -fen '8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1' \
  -depth 6 \
  -cpuprofile .codex-tmp/perft-pos3-d6.cpu.prof
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 24.83193616s
```

CPU profile inspection was then done with `pprof` in local web UI mode.

## Re-run

On the current checkout:

```bash
./scripts/bench-perft.sh
```

On a tagged benchmark version:

```bash
git switch --detach benchmark-v0
./scripts/bench-perft.sh

git switch --detach benchmark-v1
./scripts/bench-perft.sh
```

Override the benchmark target if needed:

```bash
BENCH_FEN='your fen here' BENCH_DEPTH=6 BENCH_PROFILE=.codex-tmp/custom.cpu.prof ./scripts/bench-perft.sh
```

## Results

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v0 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 24.83s | baseline |
| v1 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 15.11s | -9.72s (-39.1%) |
| v2 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 8.46s | -6.65s (-44.0%) |

## Optimization Log

### v0

Baseline only. No optimization applied yet.

Observed top CPU hotspots from the first profile:

- `chessV2/internal.scanRayForAttack`
- `chessV2/internal.(*Position).PieceAt`
- `chessV2/internal.isSameLineOrRow`

### v1

Optimizations applied:

- Replaced iterative sliding attack ray scans with bitboard-first first-blocker lookup
- Added a direct `[64]Piece` square cache on `Position` so `PieceAt` is O(1) without reconstructing from bitboards
- Stored and restored the square cache in `MoveHistory` / `UnMakeMove`
- Removed `PieceAt` and `isSameLineOrRow` from the attack-scan inner loop

Benchmark command:

```bash
./.codex-tmp/perft-profiler \
  -fen '8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1' \
  -depth 6 \
  -cpuprofile .codex-tmp/perft-pos3-d6-v1.cpu.prof
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 15.113510305s
```

### v2

Optimizations applied:

- Replaced full-position `MoveHistory` snapshots with a compact undo record
- Stored only move delta data needed for undo: moved piece, captured piece, capture square, rights, en passant, king positions, and king safety caches
- Switched `MakeMove` / `UnMakeMove` to pass `MoveHistory` by value instead of allocating and returning `*MoveHistory`
- Reworked `UnMakeMove` to restore board state incrementally instead of copying all bitboards and the full square cache
- Simplified en passant handling so make/unmake directly remove and restore the captured pawn on its real square

Benchmark command:

```bash
./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 8.463420271s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

## Update Rules

For each new version:

1. Re-run the same position and depth unless intentionally changing the benchmark suite.
2. Add a new row to the results table.
3. Record the concrete optimizations made for that version below.
4. Keep the old rows unchanged so regressions remain visible.
