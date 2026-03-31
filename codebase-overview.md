# Codebase Overview

## Purpose

This repository is a small Go chess engine focused on:

- legal move generation
- `make` / `unmake`
- attack detection
- perft-style correctness and performance benchmarking

It is not a full playing engine yet. There is no search, evaluation, time management, or full UCI game loop.

## High-Level Architecture

The current architecture is centered around five layers:

1. `Position`: mutable board state and cached state
2. pseudo-legal move generation
3. position analysis for checks / pins / evasions
4. move application and undo
5. perft recursion and optional perft-only tricks

In practice, the hot path is:

1. generate pseudo-legal targets for a piece
2. analyze the position once for checks and pins
3. filter legal moves in `Engine`
4. apply moves through `PositionUpdater`
5. recurse for perft
6. undo with `MoveHistory`

## Repository Layout

### Entrypoints

- `cmd/perft.go`
  Legacy perft entrypoint.
- `cmd/benchperft/main.go`
  Main benchmark executable used by `scripts/bench-perft.sh`.
- `cmd/main.go`
  Placeholder.

### Current packages

- `internal/chess/`
  Current chess domain package. It contains board state, move generation, legality, make/unmake, hashing, and perft orchestration.
- `internal/search/`
  Reserved for future search code.
- `internal/eval/`
  Reserved for future evaluation code.
- `internal/lichess/`
  Reserved for future Lichess integration.

### Core chess files

- `internal/chess/position.go`
  Mutable board state, FEN parsing, occupancy masks, piece boards, king caches.
- `internal/chess/position-analysis.go`
  Position-level analysis for checks, pinned pieces, and evasion masks.
- `internal/chess/pseudo-legal-move-generator.go`
  Pseudo-legal move generation and precomputed attack / ray tables.
- `internal/chess/legal-move-generator.go`
  Legal move generation and move emission on top of the pseudo-legal primitives.
- `internal/chess/engine.go`
  Engine construction plus recursive perft orchestration.
- `internal/chess/position-updater.go`
  Plain move application / undo without Zobrist maintenance.
- `internal/chess/position-updater-zobrist.go`
  Zobrist decorator layered on top of the plain updater.
- `internal/chess/check.go`
  Attack detection and `IsKingInCheck`.
- `internal/chess/move.go`
  `Move` representation.
- `internal/chess/move-history.go`
  Compact undo record used by `UnMakeMove`.
- `internal/chess/perft_tt.go`
  Perft transposition table for tricks-enabled perft.
- `internal/chess/zobrist.go`
  Zobrist table initialization and key helpers.
- `internal/chess/tools.go`
  Generic square / bitboard helpers.

### Tests

- `internal/chess/engine_legal_moves_test.go`
  Exact legal move regressions.
- `internal/chess/engine_perft_test.go`
  Perft regression counts.
- `internal/chess/position-updater_test.go`
  Make / unmake, castle rights, en passant, active color.
- `internal/chess/position-updater-zobrist_test.go`
  Plain updater vs Zobrist decorator behavior.
- `internal/chess/position-analysis_test.go`
  Direct tests for check / pin analysis.
- `internal/chess/check_test.go`
  Attack-detection helpers.
- `internal/chess/pseudo-legal-move-generator_test.go`
  Pseudo-legal move generation tests and microbenchmarks.
- `internal/chess/move_test.go`
  `Move` helper tests.
- `internal/chess/move-history_test.go`
  Packed undo-state tests.

### Scripts and docs

- `scripts/bench-perft.sh`
  Standard benchmark wrapper.
- `scripts/perft-diff.sh`
  Root-move diffing against Stockfish.
- `benchmark-history.md`
  Versioned benchmark log and optimization history.
- `future-optimisations.md`
  Remaining optimization ideas.
- `codebase-overview.md`
  Current package layout and architecture notes.

## Position Model

`Position` is a single mutable board object.

It stores:

- active color
- castling rights for both sides
- en passant square
- white and black king squares
- aggregate occupancy bitboards
- color occupancy bitboards
- per-piece-type bitboards
- board mailbox as `[64]Piece`
- cached king safety values
- cached king affect masks
- Zobrist key

Important point:

- the engine now keeps both bitboards and a `[64]Piece` mailbox
- hot code often uses the mailbox for direct piece lookup and the bitboards for attacks / occupancy

## Move Model

`Move` stores:

- moving piece
- start square
- end square
- flag

Flags currently include:

- `NormalMove`
- promotion flags
- `EnPassant`
- `PawnDoubleMove`
- `Castle`
- `Capture`

`internal/chess/move.go` owns the `Move` value and its public helpers such as `StartIdx()`, `EndIdx()`, and `UCI()`.

## Pseudo-Legal Move Generation

`PseudoLegalMoveGenerator` precomputes:

- slider rays
- knight attacks
- king attacks
- pawn push masks
- pawn capture masks

Pseudo-legal generation is split by piece family:

- `PawnPseudoLegalMovesInto`
- `KnightPseudoLegalMovesInto`
- `KingPseudoLegalMovesInto`
- `SliderPseudoLegalMovesInto`

Current style:

- generate target squares into caller-owned buffers
- let `Engine` apply legality filtering

## Position Analysis

`internal/chess/position-analysis.go` contains the intermediate analysis used by legal move generation.

`positionAnalysis` currently computes:

- whether the side to move is in check
- number of checkers
- evasion mask for single-check positions
- pinned piece mask
- per-piece pin rays

This is used by `Engine.legalMovesInto(...)` to avoid unnecessary `make` / `unmake`.

## Legal Move Generation

`Engine.LegalMoves(...)` is the public legal move entrypoint.

`legalMovesInto(...)` in `internal/chess/legal-move-generator.go` is the real hot path.

The current flow is:

1. Identify the moving side king square
2. Compute one `positionAnalysis`
3. Iterate friendly pieces from occupancy
4. Generate pseudo-legal targets into a fixed buffer
5. Filter by:
   - double-check constraints
   - single-check evasion mask
   - pin ray restrictions
   - direct king-square attack checks
6. Fall back to `MakeMove` / `UnMakeMove` only for tricky cases that still need validation

This is no longer a pure “generate everything then make/unmake everything” design.

## Move Application and Undo

There are now two move-application modes.

### Plain updater

`internal/chess/position-updater.go`

Responsible for:

- board mutation
- captures
- promotions
- castling rook movement
- en passant resolution
- en passant target updates
- castling-right updates
- king square updates
- restoring state from `MoveHistory`

`PlainPositionUpdater` does not maintain Zobrist keys.

### Zobrist decorator

`internal/chess/position-updater-zobrist.go`

Responsible for:

- wrapping the plain updater
- storing the previous Zobrist key in the undo record
- applying incremental Zobrist updates on `MakeMove`
- restoring the previous key on `UnMakeMove`

This is effectively a decorator over the plain updater.

### Constructor behavior

- `NewPositionUpdater(...)` returns the Zobrist-decorated updater
- `NewPlainPositionUpdater(...)` returns the plain updater

`Engine.SetPerftTricks(false)` switches to the plain updater because perft without TT does not need Zobrist maintenance.

## MoveHistory

`MoveHistory` is now a compact undo record, not a broad state snapshot.

It stores:

- previous Zobrist key
- previous king affect masks
- move
- captured piece
- capture square
- packed previous state

The packed state includes:

- previous king squares
- previous en passant square
- previous castle rights
- previous king safety caches

This file is intentionally more systems-level than user-facing.

## Check Detection

`internal/chess/check.go` contains:

- pawn attack checks
- knight attack checks
- slider attack checks
- general square attack helper
- `IsKingInCheck`

This code is shared by:

- king legality checks
- castling path checks
- cached king safety updates

## Perft Path

Perft orchestration lives in `internal/chess/engine.go`.

There are two relevant modes:

### Tricks off

- plain updater
- no perft transposition table
- no bulk-count caching benefits beyond the current control flow
- closest to raw movegen / make-unmake performance

### Tricks on

- Zobrist updater
- depth-2 bulk counting
- perft transposition table

This is why tricks-on and tricks-off timings should be read separately in `benchmark-history.md`.

## Current Design Strengths

- Clear split between pseudo-legal generation and legal filtering
- `Position` is compact enough to mutate in place
- direct tests exist for the refactor-sensitive helpers
- perft-only tricks are now separated from the plain updater path
- undo state is much lighter than earlier versions

## Current Design Tensions

These are the main areas to watch in future refactors.

### 1. `internal/chess` is still a broad package

The new directory structure is cleaner, but `internal/chess` still groups position state, attack analysis, legal generation, and perft orchestration in one package.
That is an intentional intermediate step before a later package split.

### 2. `positionAnalysis` is intentionally internal

It is not a method on `Position`; it is a derived analysis object used by legal move generation. That is a good separation, but it means analysis logic is spread across files rather than living directly on `Position`.

### 3. `MoveHistory` is efficient but more opaque

The packed representation is good for performance, but needs comments and tests to stay maintainable.

## Short Mental Model

If returning to this code later, the shortest useful model is:

- `Position` is the mutable board and cache container
- `PseudoLegalMoveGenerator` generates pseudo-legal targets
- `positionAnalysis` computes checks / pins once per node
- `Engine` turns pseudo-legal targets into legal `Move`s
- `PlainPositionUpdater` applies and undoes moves
- `ZobristPositionUpdater` decorates that path when hashing is needed
- perft can run with or without tricks, and the two modes are intentionally different

## Notes For Future Sessions

- Do not assume `position-updater.go` owns hashing anymore; it does not
- Do not assume `MoveHistory` is a broad snapshot; it is packed
- Do not assume the future `movegen` package split is already done; today the legal and pseudo-legal generators still live together inside `internal/chess`
- Read `benchmark-history.md` before making performance claims
