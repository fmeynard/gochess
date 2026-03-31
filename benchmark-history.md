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

Disable perft-only tricks such as bulk counting and TT:

```bash
BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
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

`benchperft` also supports `-no-perft-tricks` directly. In this repository, "perft tricks" currently means:

- bulk counting at `depth == 2`
- perft transposition-table lookups/stores

## Results

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v0 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 24.83s | baseline |
| v1 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 15.11s | -9.72s (-39.1%) |
| v2 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 8.46s | -6.65s (-44.0%) |
| v3 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 5.47s | -2.99s (-35.4%) |
| v4 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 2.28s | -3.19s (-58.3%) |
| v5 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 1.67s | -0.61s (-26.5%) |
| v6 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 229ms / 921ms | tricks on / off |

## Experimental Tiers On v5 Base

These measurements were run on top of the `benchmark-v5` code line and are not directly comparable to the earlier `v0..v5` table above. They measure the staged integration of ideas from the external `diff.txt`.

| Stage | Time | Speedup vs tier baseline | Notes |
| --- | ---: | ---: | --- |
| Tier baseline | 1.515s | 1.00x | Current `benchmark-v5` baseline before external tier experiments |
| Tier 1 | 808ms | 1.88x | Bulk counting at `depth==2`, concrete generator/updater types, fully lazy king cache invalidation |
| Tier 2 | 766ms | 1.98x | Fixed-size rook/bishop rays, corrected mask initialization, pawn attack lookup table, compact sliding attack checks |
| Tier 3 | 394ms | 3.84x | Pinned-piece bitboard fast path in `legalMovesInto` |
| Tier 4 | 206ms | 7.36x | Incremental Zobrist hashing plus perft transposition table |

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

### v3

Optimizations applied:

- Replaced generic `setPieceAt` usage in the hot make/unmake path with targeted `addPieceAt` / `removePieceAt` updates
- Changed move application to avoid clearing every piece bitboard when the moved or captured piece is already known
- Removed the direction-to-ray lookup loop from `scanRayForAttack` by passing the precomputed slider mask index directly
- Kept generic `setPieceAt` only as a slower fallback for non-hot-path board updates

Benchmark command:

```bash
./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 5.468913786s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

### v4

Optimizations applied:

- Added buffer-based pseudo-legal move generation APIs to the move generator
- Reworked the perft/search hot path to use `legalMovesInto(...)` instead of allocating a fresh `[]Move` on each recursive call
- Added fixed per-ply move storage for perft recursion with `[MaxPerftPly][MaxLegalMoves]Move`
- Removed temporary `[]Move` candidate slices from legal move filtering and wrote moves directly into caller-owned buffers
- Kept `LegalMoves()` as a compatibility wrapper that copies from the buffered path for tests and non-hot-path callers

Benchmark command:

```bash
./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 2.281543999s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

### v5

Optimizations applied:

- Changed king-safety maintenance in `MakeMove` from eager recomputation to lazy invalidation
- Kept the existing king safety cache, but now mark it `NotCalculated` when a move may affect a king and defer the actual `IsKingInCheck` work until it is needed
- Reduced immediate work inside `updateMovesAfterMove`, which was still showing up prominently after the v4 move-buffer pass
- Specialized hot board updates in `MakeMove` / `UnMakeMove` with direct `movePiece` / `capturePiece` operations
- Removed more generic remove/add sequences and unnecessary destination lookups from the move application hot path
- Added a lightweight cached `kingAffectMask` per king so `IsMoveAffectsKing` can do a single mask test instead of recomputing per-call king attack influence

Benchmark command:

```bash
./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 2.122609595s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

Updated after specializing hot board updates:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 2.044383617s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

Updated after adding cached king affect masks:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Nodes: 11030083
Elapsed: 1.674778152s
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

### v6

Optimizations applied:

- Tier 1: concrete generator/updater types, bulk counting at `depth==2`, fully lazy king safety invalidation
- Tier 2: fixed-size rook/bishop ray storage, one-time mask initialization, pawn attack lookup table, compact sliding attack checks
- Tier 3: pinned-piece bitboard fast path in `legalMovesInto`
- Tier 4: incremental Zobrist hashing and perft transposition table

Benchmark command:

```bash
./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Perft tricks: true
Nodes: 11030083
Elapsed: 228.56469ms
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

With tricks disabled:

```bash
BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Perft tricks: false
Nodes: 11030083
Elapsed: 921.093998ms
CPU profile: .codex-tmp/bench-perft.cpu.prof
```

Note: `v6` with tricks enabled includes bulk counting and a perft transposition table, so it is not a pure raw movegen/make-unmake comparison against earlier versions without TT. The `-no-perft-tricks` measurement is the closer raw-engine comparison for this codebase state.

## Update Rules

For each new version:

1. Re-run the same position and depth unless intentionally changing the benchmark suite.
2. Add a new row to the results table.
3. Record the concrete optimizations made for that version below.
4. Keep the old rows unchanged so regressions remain visible.
