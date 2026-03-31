# Codebase Overview

## Purpose

This repository is a Go chess engine currently optimized around:

- legal move generation
- `make` / `unmake`
- attack detection
- perft correctness
- perft-oriented performance profiling

It is not a full playing engine yet. Search, evaluation, self-play control, and Lichess integration are planned separately.

## Package Layout

- `internal/board`
  Owns the mutable board state and its mutation primitives.
- `internal/movegen`
  Owns pseudo-legal generation, legal filtering, and attack/check analysis.
- `internal/engine`
  Owns orchestration: public engine API, perft recursion, and perft TT wiring.
- `internal/search`
  Owns search types and future search logic.
- `internal/eval`
  Owns score semantics and future static evaluation.
- `internal/lichess`
  Reserved for future integration with Lichess.

## Mental Model

- `board.Position` is the source of truth for board state.
- `board.Move` and `board.MoveHistory` are the move and undo records.
- `board.MoveApplier` applies and undoes moves.
- `movegen.PseudoLegalMoveGenerator` owns both pseudo-legal helpers and legal move generation.
- `engine.Engine` wires `movegen`, `board`, `eval`, and `search` together for public use.

The dependency direction is intentional:

1. `board`
2. `movegen -> board`
3. `engine -> board + movegen`

That keeps the mutable board layer independent from the higher-level generation/orchestration layers.

## Entrypoints

- `cmd/perft.go`
  Simple perft divide entrypoint.
- `cmd/benchperft/main.go`
  Benchmark entrypoint used by `scripts/bench-perft.sh`.
- `cmd/main.go`
  Placeholder.

## Board Layer

Important files:

- `internal/board/position.go`
- `internal/board/move.go`
- `internal/board/move-history.go`
- `internal/board/position-updater.go`
- `internal/board/position-updater-zobrist.go`
- `internal/board/zobrist.go`
- `internal/board/tools.go`

`Position` stores:

- side to move
- castling rights
- en passant square
- king squares
- occupancy bitboards
- per-piece bitboards
- mailbox board as `[64]Piece`
- king-safety caches
- Zobrist key

The updater layer is split in two:

- `PlainPositionUpdater`
  Pure board mutation and undo
- `ZobristPositionUpdater`
  Decorator that adds incremental Zobrist maintenance

`NewPositionUpdater()` returns the Zobrist-decorated updater.
`NewPlainPositionUpdater()` returns the plain updater.

## Move Generation Layer

Important files:

- `internal/movegen/pseudo-legal-move-generator.go`
- `internal/movegen/legal-move-generator.go`
- `internal/movegen/position-analysis.go`
- `internal/movegen/check.go`
- `internal/movegen/magic_bitboards.go`

Responsibilities:

- precomputed attack masks and slider lookup tables
- pseudo-legal destination generation
- legal filtering
- check detection
- pin detection
- evasion-mask construction

The legal path is centered on `PseudoLegalMoveGenerator.LegalMovesInto(...)`.

High-level flow:

1. identify the moving side and king square
2. compute one `positionAnalysis`
3. emit king moves first
4. dispatch non-king generation through specialized legal paths
5. use `MakeMove` / `UnMakeMove` only for cases that still need dynamic validation, notably en passant legality

## Engine Layer

Important files:

- `internal/engine/engine.go`
- `internal/engine/perft_tt.go`

Responsibilities:

- public `Engine` construction
- `LegalMoves(...)`
- future `BestMoveDepth(...)` / `BestMoveTime(...)` search entrypoints
- `PerftDivide(...)`
- recursive perft traversal
- optional perft TT and depth-2 bulk counting when tricks are enabled

Perft has two meaningful modes:

- tricks off
  closest to raw movegen + updater performance
- tricks on
  adds Zobrist-backed TT and bulk-count shortcuts

Read benchmark results with that distinction in mind.

## Test Layout

Tests stay next to the code they validate:

- `internal/board/*_test.go`
- `internal/movegen/*_test.go`
- `internal/engine/*_test.go`

That is the standard Go layout and is preferable to separate `tests/` subdirectories here.

## Current Strengths

- package boundaries are now explicit
- board mutation is isolated from move generation
- legal generation is no longer mixed into engine orchestration
- search/eval scaffolding can now evolve without mixing search concepts into `movegen` or `board`
- updater and movegen hot paths have direct regression coverage
- benchmark workflow is simple and repeatable

## Current Tensions

- `movegen` still combines pseudo-legal and legal generation in one package
- `MoveHistory` is optimized and compact, so readability depends on comments/tests
- the code is still perft-driven, which is useful but not a substitute for future search benchmarks

## Related Docs

- `docs/benchmark-history.md`
- `docs/benchmark-learnings.md`
- `docs/future-optimisations.md`
