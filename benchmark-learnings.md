# Benchmark Learnings

This file records practical lessons from the perft optimization work so future sessions can reuse the context quickly and avoid retrying ideas that already regressed or introduced correctness bugs.

## Current Best Known Result

- Best raw benchmark so far: `benchmark-v12`
- Target: `Perft position 3`
- FEN: `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1`
- Mode: `BENCH_NO_PERFT_TRICKS=1`
- Depth: `7`
- Nodes: `178,633,661`
- Time: `7.08912363s`

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

## What Did Not Pay

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

Avoid retrying first:

1. Splitting the generic move append loop into separate quiet/capture helper families.
2. Replacing board-update switch helpers with pointer indirection.

Successful direction to revisit later:

1. Small, targeted reductions in attack-detection branching.
2. Small, targeted reductions in position-analysis branching.
