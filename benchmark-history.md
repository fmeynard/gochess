# Benchmark History

This file tracks repeatable performance measurements for the engine and the optimizations introduced between versions.

See also:

- `benchmark-learnings.md` for qualitative lessons, failed experiments, and correctness pitfalls discovered during the benchmark work
- `plan.md` for the current optimization target and near-term work order

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

Benchmark mode defaults to `hot`, which excludes FEN parsing and engine setup from the timed section and profiles only the perft run itself. This is now the preferred mode for movegen optimization work.

To include initialization and FEN parsing in the timed section:

```bash
BENCH_MODE=cold BENCH_WARMUP=0 ./scripts/bench-perft.sh
```

To run the old-style end-to-end timing explicitly, use `cold`. To compare movegen-only performance, use `hot`.

`benchperft` also supports `-no-perft-tricks` directly. In this repository, "perft tricks" currently means:

- bulk counting at `depth == 2`
- perft transposition-table lookups/stores

## Results

### Tricks Off

These results are the closest raw movegen / make-unmake comparison across versions.

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v0 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 24.83s | baseline |
| v1 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 15.11s | -9.72s (-39.1%) |
| v2 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 8.46s | -6.65s (-44.0%) |
| v3 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 5.47s | -2.99s (-35.4%) |
| v4 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 2.28s | -3.19s (-58.3%) |
| v5 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 1.67s | -0.61s (-26.5%) |
| v6 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 921ms | -0.75s (-45.0%) |
| v7 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 809ms | -112ms (-12.2%) |
| v8 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 776ms | -33ms (-4.1%) |
| v9 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 576ms | -200ms (-25.8%) |

### Tricks On

These results include perft-only tricks such as bulk counting and TT. `v6` is the first baseline in this category.

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v6 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 229ms | baseline |
| v7 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 256ms | +27ms (+11.8%) |
| v8 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 68ms | -188ms (-73.4%) |
| v9 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 91ms | +23ms (+33.8%) |

Note: `v9` remains focused on the raw `-no-perft-tricks` path. Tricks-on depth-6 timings are still noisy under `go run`; the single-shot `v9` result above was `90.61035ms`, while a local sample-of-4 taken on March 31, 2026 produced `303ms`, `86ms`, `86ms`, `309ms`.

## Perft(7) Snapshot

These depth-7 numbers were taken on the same benchmark FEN on March 31, 2026. They are useful for checking whether an optimization still helps once the tree is much larger.

### Tricks Off

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v8 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 9.83s | baseline |
| v9 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 9.16s | -0.67s (-6.8%) |
| v10 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 8.95s | -0.21s (-2.3%) |
| v11 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 8.09s | -0.86s (-9.6%) |
| v12 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 7.09s | -1.00s (-12.3%) |
| v13 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 6.97s | -0.12s (-1.7%) |
| v14 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 6.86s | -0.11s (-1.6%) |

### Tricks On

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v8 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 895ms | baseline |
| v9 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 715ms | -180ms (-20.1%) |

## Perft(7) Hot Benchmark

These numbers use the corrected `hot` harness, which excludes FEN parsing and engine construction from the timed section and runs on the compiled `benchperft` binary instead of `go run`.

### Tricks Off

The first `hot` baseline was taken on `benchmark-v14`. Because the harness changed, these numbers should not be mixed directly with the older `Perft(7) Snapshot` table above.

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v14 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 6.78s | hot baseline |
| v15 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 6.50s | -0.28s (-4.1%) |
| v16 | 2026-03-31 | Perft position 3 | 7 | 178,633,661 | 6.30s | -0.20s (-3.1%) |

Recorded samples:

- `v14`: `6.822986189s`, `6.745086913s`
- `v15`: `6.533347471s`, `6.462752555s`
- `v16`: `6.319541638s`, `6.295790877s`

Note: `v7` timings were noisy in single-shot runs because `./scripts/bench-perft.sh` uses `go run`. The recorded `v7` numbers above are medians from four local samples taken on March 31, 2026:

- Tricks on: `252ms`, `260ms`, `69ms`, `277ms` -> recorded median `256ms`
- Tricks off: `858ms`, `780ms`, `625ms`, `838ms` -> recorded median `809ms`

Note: `v8` was measured with the same `go run`-based script and the same median-of-4 convention:

- Tricks on: `256ms`, `67ms`, `66ms`, `70ms` -> recorded median `68ms`
- Tricks off: `802ms`, `762ms`, `790ms`, `629ms` -> recorded median `776ms`

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

### v7

Optimizations applied:

- Reworked `legalMovesInto` around a single `computePosInfo` pass that identifies checkers, evasion masks, and pinned pieces up front
- Replaced king move legality make/unmake checks with direct `isSquareAttacked(...)` tests, including castling path validation
- Added fast filtering for double-check, single-check evasion, and pinned-piece move constraints before falling back to make/unmake
- Cached `depth == 2` perft results in the transposition table instead of bypassing TT on the bulk-count fast path
- Replaced the perft TT hash map with a fixed-size direct-addressed table keyed by `zobrist ^ (depth << 56)`
- Removed `fmt.Sprintf` from `Move.UCI()` and built the string directly into a fixed buffer

Benchmark commands:

```bash
./scripts/bench-perft.sh
BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded local samples:

```text
Perft tricks: true
Elapsed: 252.221302ms
Elapsed: 260.006713ms
Elapsed: 69.270333ms
Elapsed: 276.845562ms

Perft tricks: false
Elapsed: 857.928135ms
Elapsed: 780.275579ms
Elapsed: 624.968179ms
Elapsed: 838.292528ms
```

Interpretation:

- With tricks disabled, `v7` is measurably faster than `v6`, which matches the intended reduction in make/unmake work during legal move generation
- With tricks enabled, the median result is slightly slower than `v6`; the direct-addressed TT and new legal-move fast paths do not improve this benchmark enough to offset that on this workload

### v8

Optimizations applied:

- Classified generated moves with richer `Move` flags so `MakeMove` no longer has to rediscover castling, en passant, captures, and pawn double pushes from board state in the hot path
- Compacted `MoveHistory` by packing prior king squares, castle rights, en passant, and king safety caches into a small `meta` field instead of carrying many separate fields
- Stored and restored cached king-affect masks directly in the undo record so `UnMakeMove` no longer recomputes them per node
- Disabled incremental Zobrist maintenance when perft tricks are turned off, since the hash is only needed for the perft TT in this benchmark mode
- Switched `MakeMove` / `UnMakeMove` to trust the internally generated move's `piece` payload instead of rereading it from the board in the recursive hot path

Benchmark commands:

```bash
./scripts/bench-perft.sh
BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded local samples:

```text
Perft tricks: true
Elapsed: 256.336512ms
Elapsed: 67.063821ms
Elapsed: 66.125952ms
Elapsed: 69.747462ms

Perft tricks: false
Elapsed: 802.044841ms
Elapsed: 761.603425ms
Elapsed: 789.776057ms
Elapsed: 629.293308ms
```

Interpretation:

- With tricks disabled, `v8` is a modest but real improvement over `v7`; most of the gain comes from lighter undo work and skipping unused Zobrist updates
- With tricks enabled, `v8` is much faster on warmed runs; the move-flag and undo-path cleanup materially reduces the remaining overhead around the TT-assisted perft path

### v9

Optimizations applied:

- Replaced pinned-piece ray lookup in `positionAnalysis` with direct `[64]uint64` square-indexed storage instead of scanning a compact side table
- Simplified `legalMovesInto(...)` so the inner loop reuses `enPassantIdx`, `targetMask`, and `targetPiece` rather than recomputing them across multiple branches
- Materialized `Move` values directly in the legal-move hot path instead of calling `NewMove` / `classifyMove` for every surviving target
- Reduced undo overhead in `UnMakeMove` by decoding `packedState` directly instead of going through multiple tiny accessors
- Switched the plain updater hot path to use direct move fields and mailbox lookups in `MakeMove` / `UnMakeMove`

Benchmark commands:

```bash
./scripts/bench-perft.sh
BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
BENCH_DEPTH=7 ./scripts/bench-perft.sh
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded outputs:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Perft tricks: true
Nodes: 11030083
Elapsed: 90.61035ms
CPU profile: .codex-tmp/bench-perft-v9.cpu.prof

FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 6
Perft tricks: false
Nodes: 11030083
Elapsed: 575.536727ms
CPU profile: .codex-tmp/bench-perft-no-tricks-v9-d6.cpu.prof

FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: true
Nodes: 178633661
Elapsed: 715.189908ms
CPU profile: .codex-tmp/bench-perft-v9-d7.cpu.prof

FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 9.160036102s
CPU profile: .codex-tmp/bench-perft-no-tricks-v9.cpu.prof
```

Interpretation:

- With tricks disabled, `v9` is a clear improvement over `v8` at both depth 6 and depth 7; the gain comes from less work inside `legalMovesInto(...)` and cheaper undo-state restoration
- With tricks enabled, `v9` improves the depth-7 run but the depth-6 `go run` path is still too noisy to read as a clean regression or improvement signal

### v10

Optimizations applied:

- Added a fast path in `MakeMove` for ordinary quiet moves that updates occupancies, piece boards, and mailbox entries directly without routing through `Position.movePiece(...)`
- Added a second fast path in `MakeMove` for ordinary captures that updates the mover and captured piece bitboards directly
- Added matching fast paths in `UnMakeMove` for ordinary quiet moves and ordinary captures, keeping the fallback path only for promotions, en passant, and castling
- Inlined the `packedState` assembly inside `MakeMove` so the hot path no longer calls `packMoveHistoryMeta(...)`

Benchmark commands:

```bash
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 8.946131173s
CPU profile: .codex-tmp/bench-perft-no-tricks-v10.cpu.prof
```

Interpretation:

- `v10` keeps the benchmark focus strictly on the no-tricks path and pushes the raw depth-7 benchmark below `9s`
- The remaining dominant costs are still `MakeMove`, `UnMakeMove`, and `legalMovesInto(...)`, so the next likely wins are either more specialization of the updater hot path or a larger legal-move generation refactor

### v11

Optimizations applied:

- Reworked `legalMovesInto(...)` into specialized paths per piece family instead of the previous generic “pseudo-target list then branch per target” loop
- Added direct king-move generation with attack-mask validation and castling checks up front
- Added bitboard-based knight, slider, and pawn target generation/filtering so check evasion and pin constraints are applied as masks before move materialization
- Restricted `MakeMove` / `UnMakeMove` legality validation to the remaining ambiguous case: en passant
- Added a dedicated `sliderTargetsMask(...)` helper to build rook/bishop/queen targets from rays and first-blocker lookups instead of walking a temporary target list

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 8.085356627s
CPU profile: .codex-tmp/bench-perft-no-tricks-v11.cpu.prof
```

Interpretation:

- `v11` is the first version where the heavy legal-move refactor lands cleanly and produces a large depth-7 no-tricks gain
- The profile now shows `legalMovesInto(...)` much lower than before; the raw move application path is once again the main limit

### v12

Optimizations applied:

- Reworked `MakeMove` / `UnMakeMove` to dispatch directly on `move.flag` instead of rediscovering en passant and castling from board state
- Added dedicated specialized updater paths for:
  - quiet moves
  - captures
  - en passant
  - promotions
  - castling
- Replaced repeated per-piece switch blocks in the hot path with compact board-update helpers for xor/set/clear on the piece bitboards
- Removed `decodeKingSafety(...)` from the undo hot path by restoring cached king safety directly from the packed bits
- Added explicit regression coverage for promotion undo state restoration so deep perft catches mailbox / occupancy mismatches immediately

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 7.08912363s
CPU profile: .codex-tmp/bench-perft-no-tricks-v12.cpu.prof
```

Interpretation:

- `v12` is a large raw updater win on top of `v11`, cutting another full second from the depth-7 no-tricks benchmark
- `MakeMove` and `UnMakeMove` still dominate, but they now account for materially less total runtime and the movegen refactor from `v11` remains intact

### v13

Optimizations applied:

- Simplified attacker-color branching in `isSquareAttacked(...)` so pawn attack selection and attacker occupancy are chosen once up front
- Unrolled the sliding-piece attack checks inside `isSquareAttacked(...)` to remove the small direction loop and repeated array indexing
- Simplified the pawn-attack preselection in `computePositionAnalysis(...)`
- Reworked the sliding-check / pin analysis loop in `computePositionAnalysis(...)` into direct per-direction processing to reduce repeated branching in the hot path

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 6.966829391s
CPU profile: .codex-tmp/bench-perft-no-tricks-v13b.cpu.prof
```

Interpretation:

- `v13` improves the attack-detection and position-analysis side enough to push the raw depth-7 benchmark under `7s`
- The next bottleneck remains the move updater, but the movegen/attack side is now materially cheaper than in `v12`

### v14

Optimizations applied:

- Added focused updater microbenchmarks covering make/unmake for:
  - pawn quiet
  - pawn double
  - pawn capture
  - knight quiet
  - rook quiet
  - king quiet
  - castling
  - promotion
- Specialized the ordinary quiet/capture updater paths further for the most common piece families (`pawn`, `rook`, `king`) instead of always routing through the generic bitboard switch helper
- Kept the rare paths unchanged so the optimization stays narrow and measurable

Benchmark commands:

```bash
BENCH_DEPTH=7 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
GOCACHE=/home/fab/Projects/gochess/.codex-tmp/go-build-cache go test ./internal -run '^$' -bench 'BenchmarkPositionUpdaterMakeUnmake' -benchtime=200ms
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Perft tricks: false
Nodes: 178633661
Elapsed: 6.857738422s
CPU profile: .codex-tmp/bench-perft-no-tricks-v14.cpu.prof

BenchmarkPositionUpdaterMakeUnmakePawnQuiet-16        43.17 ns/op
BenchmarkPositionUpdaterMakeUnmakePawnDouble-16       77.41 ns/op
BenchmarkPositionUpdaterMakeUnmakePawnCapture-16      49.84 ns/op
BenchmarkPositionUpdaterMakeUnmakeKnightQuiet-16      56.64 ns/op
BenchmarkPositionUpdaterMakeUnmakeRookQuiet-16        66.83 ns/op
BenchmarkPositionUpdaterMakeUnmakeKingQuiet-16        40.55 ns/op
BenchmarkPositionUpdaterMakeUnmakeCastle-16           45.25 ns/op
BenchmarkPositionUpdaterMakeUnmakePromotion-16        42.21 ns/op
```

Interpretation:

- `v14` is a modest but real updater-only win on top of `v13`
- The microbenchmarks are now available to guide future updater work instead of relying only on full perft timings

## Update Rules

For each new version:

1. Re-run the same position and depth unless intentionally changing the benchmark suite.
2. Add a new row to the results table.
3. Record the concrete optimizations made for that version below.
4. Keep the old rows unchanged so regressions remain visible.
