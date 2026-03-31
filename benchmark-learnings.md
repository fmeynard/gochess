# Benchmark Learnings

This file records practical lessons from the perft optimization work so future sessions can reuse the context quickly and avoid retrying ideas that already regressed or introduced correctness bugs.

## Current Best Known Result

- Best raw benchmark so far: `benchmark-v14`
- Best hot benchmark so far: `benchmark-v15`
- Target: `Perft position 3`
- FEN: `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1`
- Mode: `BENCH_NO_PERFT_TRICKS=1`
- Depth: `7`
- Nodes: `178,633,661`
- Time: `6.857738422s`

Current preferred reference:

- Harness: `hot`
- Samples: `6.533347471s`, `6.462752555s`
- Recorded reference: `6.50s`

## Strategy Context

- `perft` is a validation and profiling tool for move generation, not the long-term design center of the engine.
- The medium-term product goal is still a real chess engine:
  - legal move generation
  - scoring / search
  - engine-vs-engine play
  - later UCI / Lichess integration
- Because of that, prefer optimizations that also help a future searcher:
  - faster attack detection
  - faster sliders
  - cheaper legal filtering
  - cheaper and robust `MakeMove` / `UnMakeMove`
- Be more cautious with optimizations that create a separate perft-only architecture or make the updater harder to trust.
- Current short-term target remains aggressive: get `perft(7)` below `1s` on the benchmark FEN without perft tricks.
- The benchmark harness should optimize for measuring move generation, not one-time setup:
  - prefer `hot` mode
  - exclude FEN parsing and engine construction from the timed section
  - use the compiled `benchperft` binary rather than `go run`

## What Worked

### v9

- Replacing the pinned-piece side table scan with direct `pinRayBySq[64]` storage helped.
- Simplifying `legalMovesInto(...)` inner-loop bookkeeping helped.
- Decoding packed undo state directly in `UnMakeMove` helped.

Takeaway:

- Cheap constant-factor cuts in legal filtering and undo restoration were worth doing before larger refactors.

### v10

- Specializing the no-tricks updater for ordinary quiet moves and ordinary captures helped.
- Keeping fallback paths only for en passant, promotion, and castling was a good tradeoff.

Takeaway:

- Raw make/unmake cost was still dominating after the first movegen improvements, so updater specialization had strong ROI.

### v11

- The heavy legal move generation refactor paid off.
- Specialized paths for king, pawn, knight, and slider handling reduced `legalMovesInto(...)` materially.
- Restricting `MakeMove` / `UnMakeMove` validation to en passant only was a large win.

Takeaway:

- The old generic “generate pseudo targets then branch per target” pattern had reached its limit.

### v12

- Dispatching updater behavior directly from `move.flag` helped.
- Dedicated fast paths for quiet move, capture, en passant, promotion, and castling helped.
- Removing `isEnPassantMove`, `isCastleMove`, and `decodeKingSafety` from the updater hot path helped.

Takeaway:

- Once move generation was cheaper, eliminating hot-path rediscovery in the updater produced another large gain.

### v13

- Optimizing `computePositionAnalysis(...)` and `isSquareAttacked(...)` paid.
- Simplifying attacker-color setup and unrolling sliding attack checks helped.
- Direct per-direction processing in position analysis helped without requiring a larger architectural refactor.

Takeaway:

- After `v12`, attack detection and king-safety analysis were the right place to look, and small structural simplifications there still had measurable value.

### v14

- Adding updater microbenchmarks was worthwhile.
- A narrow specialization of the ordinary quiet/capture updater path for `pawn`, `rook`, and `king` still paid.
- This was much safer than another broad updater rewrite because the benchmark harness made the tradeoff obvious immediately.

Takeaway:

- For updater work, a “microbenchmark first, then narrow specialization” workflow is effective in this repository.
- The benchmark can still move materially with careful hot-path specialization, but broad structural rewrites are no longer automatically wins.

### v15

- Replacing slider first-blocker scans with magic-bitboard rook/bishop attack lookups paid under the corrected `hot` benchmark.
- The same magic lookups helped both legal move generation and attack detection.
- The corrected benchmark harness made the improvement much easier to judge than the older `go run`-based timing.

Takeaway:

- Magic-style slider indexing is a promising direction in this codebase when the index computation is just `(occ&mask)*magic >> shift`.
- Reusing the same slider lookup path across movegen and attack checks is a good multiplier.

## What Did Not Pay

### v15 attempt: broad 4-axis experiment

- Combining these ideas in one pass did not produce a real win:
  - more mask-first move materialization
  - `leaf-safe/full` updater split
  - table-driven slider lookups with compressed occupancy indexing
  - broader movegen cleanup around those changes
- Under the corrected `hot` benchmark, the strongest surviving variant was still not better than `v14`.

Takeaway:

- Do not bundle these axes together again.
- Evaluate them independently under the `hot` benchmark harness.

### v13 attempt: movegen split into quiet/capture append helpers

- Splitting `appendMovesFromMask(...)` into separate quiet/capture/pawn-push helpers passed tests after fixes, but regressed the benchmark.
- Measured result:
  - `v12`: `7.08912363s`
  - `v13` attempt: `7.224311853s`

Takeaway:

- Extra helper structure and mask splitting in this area cost more than they saved.
- Do not retry this exact refactor without a different profiling hypothesis.

### v13 attempt: updater piece-board pointer indirection

- Replacing `xorPieceBoard` / `clearPieceBoard` / `setPieceBoard` calls with `pieceBoardPtr(...)` pointer selection did not pay.
- The extra pointer selection and code shape appear worse than the simple switch helpers.

Takeaway:

- For this codebase and compiler, the compact switch-based board helpers are preferable to pointer indirection.

### v15 attempt: compressed slider lookup tables

- The first slider table-driven implementation regressed badly.
- `sliderLookupIndex(...)` itself became a major hotspot in `pprof`.

Takeaway:

- A lookup-table slider approach is still worth exploring, but not with the current bit-compression/indexing cost.
- If revisited, prefer a design with cheaper indexing, such as a real magic-bitboard style lookup.

### v15 attempt: `leaf-safe/full` updater split via hot-path predicate

- The updater split regressed when the code had to classify the move path dynamically in the hot path.
- `isSimpleMovePath(...)` became visible in the profile.

Takeaway:

- If this idea is retried, the path selection must come essentially for free.
- The likely workable version is to derive the fast-path category directly from move generation or move flags, not by reclassifying inside `MakeMove` / `UnMakeMove`.

## Correctness Pitfalls Already Hit

### Promotion undo bug

Symptom:

- Deep perft node count collapsed while small regression tests still passed.

Cause:

- Undoing a non-capturing promotion restored piece/color occupancy but failed to clear `pos.occupied` on the promotion square.

Protection:

- `internal/position-updater_test.go` now contains explicit regression coverage for promotion undo with and without capture.

### King capture semantics

Symptom:

- Historical tests/perft counts implicitly allowed move generation to land on the enemy king square.

Correction:

- Legal move generation now forbids moves whose destination is the enemy king square.
- Promotion regression expectations were updated accordingly.

Protection:

- `internal/engine_legal_moves_test.go` now checks that legal moves never capture the enemy king.
- `internal/engine_perft_test.go` contains a dedicated regression for the promotion/perft case.

## Recommended Next Areas

Based on the `v12` profile:

- `MakeMove`
- `UnMakeMove`
- `legalMovesInto`
- `appendKingMoves`
- `appendPawnMoves`
- `computePositionAnalysis`
- `isSquareAttacked`

Preferred next experiments:

1. Optimize `appendKingMoves(...)` without adding more helper fragmentation.
2. Optimize `appendPawnMoves(...)` with careful side-specific specialization.
3. Look for ways to share or cheapen attack computations between `computePositionAnalysis(...)` and `isSquareAttacked(...)`.
4. Use the new updater microbenchmarks before making further `MakeMove` / `UnMakeMove` changes.
5. Explore slider attack generation upgrades that are useful beyond perft, including denser lookup schemes or magic-style indexing.
6. Reduce move materialization overhead in legal generation without splitting the code into many tiny helpers.
7. Extend magic-based attacks into `computePositionAnalysis(...)` so check and pin discovery also stop paying repeated directional scans.

## Architectural Guardrails

- Keep the engine core usable for future search code.
- Avoid introducing a separate perft-only move legality model unless the gain is overwhelming and the maintenance cost is explicit.
- Favor changes that would still make sense under alpha-beta:
  - stable move representation
  - robust make/unmake
  - clear attack / check logic
  - cache-friendly board state
- If a benchmark-only optimization is considered, document it clearly in `benchmark-history.md` and keep it isolated.

Avoid retrying first:

1. Splitting the generic move append loop into separate quiet/capture helper families.
2. Replacing board-update switch helpers with pointer indirection.

Successful direction to revisit later:

1. Small, targeted reductions in attack-detection branching.
2. Small, targeted reductions in position-analysis branching.
