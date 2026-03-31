# Benchmark History

This file tracks the current benchmark reference for engine optimization work, plus the older benchmark suites kept for historical comparison.

See also:

- `benchmark-learnings.md` for qualitative lessons, failed experiments, and correctness pitfalls
- `plan.md` for the current optimization target and near-term work order

## Methodology

The current reference benchmark is:

- target: raw move generation / legality / make-unmake throughput
- position label: `Perft position 3`
- FEN: `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1`
- depth: `7`
- mode: `hot`
- perft tricks: `off`
- runner: compiled `benchperft` binary

Why `hot`:

- it excludes FEN parsing and engine construction from the timed section
- it focuses the measurement on the recursive engine path that actually matters for movegen optimization
- one-time setup cost is not a useful primary target here

Why `no-tricks`:

- it disables bulk counting and the perft TT
- it keeps the benchmark focused on the real engine path rather than perft-specific shortcuts
- it is a better proxy for future search-time move generation cost

## Re-run

Current reference run:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Cold end-to-end comparison, only if setup cost matters for a specific experiment:

```bash
BENCH_DEPTH=7 BENCH_MODE=cold BENCH_WARMUP=0 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

On a tagged benchmark version:

```bash
git switch --detach benchmark-v16
./scripts/bench-perft.sh

git switch --detach benchmark-v15
./scripts/bench-perft.sh
```

Override the target if needed:

```bash
BENCH_FEN='your fen here' BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 BENCH_PROFILE=.codex-tmp/custom.cpu.prof ./scripts/bench-perft.sh
```

In this repository, "perft tricks" currently means:

- bulk counting at `depth == 2`
- perft transposition-table lookups/stores

## Current Results

Current reference suite:

- `perft(7)`
- `hot`
- `no-tricks`

Older benchmark suites are kept in [Historical Results](#historical-results).

| Version | Date | Position | Depth | Mode | Tricks | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | --- | --- | ---: | ---: | ---: |
| [v14](#current-v14) | 2026-03-31 | Perft position 3 | 7 | hot | off | 178,633,661 | 6.78s | hot baseline |
| [v15](#current-v15) | 2026-03-31 | Perft position 3 | 7 | hot | off | 178,633,661 | 6.50s | -0.28s (-4.1%) |
| [v16](#current-v16) | 2026-03-31 | Perft position 3 | 7 | hot | off | 178,633,661 | 6.30s | -0.20s (-3.1%) |
| [v17](#current-v17) | 2026-03-31 | Perft position 3 | 7 | hot | off | 178,633,661 | 6.21s | -0.09s (-1.4%) |

Recorded samples:

- `v14`: `6.822986189s`, `6.745086913s`
- `v15`: `6.533347471s`, `6.462752555s`
- `v16`: `6.319541638s`, `6.295790877s`
- `v17`: `6.190472756s`, `6.22881615s`

## Optimization Log

<a id="current-v17"></a>
### v17

Optimizations applied:

- specialized `legalMovesInto(...)` into separate non-king paths for:
  - `no-check && no-pins`
  - `no-check`
  - `in-check`
- added a dedicated `appendNonKingMovesNoCheckNoPins(...)` fast path so the common legal-generation case avoids repeated check/pin branching in the inner loop
- kept move materialization unchanged after earlier `v17` materialization splits failed to beat `v16`

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Mode: hot
Warmup: 1
Perft tricks: false
Nodes: 178633661
Elapsed: 6.22881615s
CPU profile: .codex-tmp/bench-perft-no-tricks-hot-v17c-rerun.cpu.prof
```

Interpretation:

- `v17` is a smaller but real win on top of `v16`
- reducing legal-generation branching before move materialization paid, while changing move materialization itself still did not
<a id="current-v16"></a>
### v16

Optimizations applied:

- reworked `computePositionAnalysis(...)` so check and pin discovery iterates candidate pinners/checkers through precomputed `betweenMasks` instead of repeated directional first-blocker scans
- added `betweenMasks[from][to]` and `orthogonalAttacksMask` precomputations to support the new analysis path
- reused the magic-slider infrastructure from `v15` as the candidate generator for rook/queen and bishop/queen attackers

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Mode: hot
Warmup: 1
Perft tricks: false
Nodes: 178633661
Elapsed: 6.295790877s
CPU profile: .codex-tmp/bench-perft-no-tricks-hot-v16-final-seq.cpu.prof
```

Interpretation:

- `v16` is a real follow-up win on top of `v15`
- once slider lookups were cheap, removing repeated directional scans from pin/check analysis paid immediately

<a id="current-v15"></a>
### v15

Optimizations applied:

- replaced rook/bishop slider first-blocker scans with magic-bitboard attack lookups
- reused the same magic slider lookups in both legal move generation and attack detection
- added one-shot attack table initialization so the compiled benchmark binary can use the same lookup tables consistently in `hot` mode

Benchmark command:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Mode: hot
Warmup: 1
Perft tricks: false
Nodes: 178633661
Elapsed: 6.462752555s
CPU profile: .codex-tmp/bench-perft-no-tricks-hot-magic-b.cpu.prof
```

Interpretation:

- `v15` is the first version benchmarked primarily under the corrected `hot` harness
- magic slider indexing produced a clear win once the index computation was reduced to a simple `(occ&mask)*magic >> shift`

<a id="current-v14"></a>
### v14

Optimizations applied:

- added focused updater microbenchmarks covering make/unmake for:
  - pawn quiet
  - pawn double
  - pawn capture
  - knight quiet
  - rook quiet
  - king quiet
  - castling
  - promotion
- specialized the ordinary quiet/capture updater paths further for the most common piece families (`pawn`, `rook`, `king`)
- kept the rare paths unchanged so the optimization stayed narrow and measurable

Benchmark commands:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
GOCACHE=/home/fab/Projects/gochess/.codex-tmp/go-build-cache go test ./internal -run '^$' -bench 'BenchmarkPositionUpdaterMakeUnmake' -benchtime=200ms
```

Recorded output:

```text
FEN: 8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1
Depth: 7
Mode: hot
Warmup: 1
Perft tricks: false
Nodes: 178633661
Elapsed: 6.745086913s

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

- `v14` is the last current-reference version before the benchmark harness pivot to `hot`
- the updater microbenchmarks remain useful for later updater work even though the main results table is now `hot` only

## Historical Results

This section keeps the older benchmark suites for context. The current optimization target is the [Current Results](#current-results) table above.

### Historical Tables

#### Depth-6 Tricks Off

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

#### Depth-6 Tricks On

| Version | Date | Position | Depth | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | ---: | ---: | ---: |
| v6 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 229ms | baseline |
| v7 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 256ms | +27ms (+11.8%) |
| v8 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 68ms | -188ms (-73.4%) |
| v9 | 2026-03-31 | Perft position 3 | 6 | 11,030,083 | 91ms | +23ms (+33.8%) |

#### Historical Perft(7)

| Version | Date | Position | Depth | Mode | Tricks | Nodes | Time | Delta vs previous |
| --- | --- | --- | --- | --- | --- | ---: | ---: | ---: |
| v8 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 9.83s | baseline |
| v9 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 9.16s | -0.67s (-6.8%) |
| v10 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 8.95s | -0.21s (-2.3%) |
| v11 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 8.09s | -0.86s (-9.6%) |
| v12 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 7.09s | -1.00s (-12.3%) |
| v13 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 6.97s | -0.12s (-1.7%) |
| v14 | 2026-03-31 | Perft position 3 | 7 | pre-hot | off | 178,633,661 | 6.86s | -0.11s (-1.6%) |
| v8 | 2026-03-31 | Perft position 3 | 7 | pre-hot | on | 178,633,661 | 895ms | tricks-on baseline |
| v9 | 2026-03-31 | Perft position 3 | 7 | pre-hot | on | 178,633,661 | 715ms | -180ms (-20.1%) |

#### Historical Notes

`v7` and `v8` were still measured through `go run`, so they were noisy enough that medians from several local samples were more useful than single-shot timings.

#### Experimental Tiers On v5 Base

| Stage | Time | Speedup vs tier baseline | Notes |
| --- | ---: | ---: | --- |
| Tier baseline | 1.515s | 1.00x | Current `benchmark-v5` baseline before external tier experiments |
| Tier 1 | 808ms | 1.88x | Bulk counting at `depth==2`, concrete generator/updater types, fully lazy king cache invalidation |
| Tier 2 | 766ms | 1.98x | Fixed-size rook/bishop rays, corrected mask initialization, pawn attack lookup table, compact sliding attack checks |
| Tier 3 | 394ms | 3.84x | Pinned-piece bitboard fast path in `legalMovesInto` |
| Tier 4 | 206ms | 7.36x | Incremental Zobrist hashing plus perft transposition table |

## Update Rules

For each new benchmark version:

1. Re-run the current reference suite unless intentionally changing the benchmark spec:
   `BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh`
2. Add a new row to `Current Results`, with the version linking to its optimization log entry.
3. Add the new optimization log entry directly under `Optimization Log`, at the top, so the log stays in reverse chronological order.
4. Keep the current-results table limited to the active `perft(7)` / `hot` / `no-tricks` suite.
5. If a benchmark belongs to an older or different suite, record it in `Historical Results` instead of the current table.
6. Preserve old historical rows so regressions and benchmark-harness changes remain visible.
