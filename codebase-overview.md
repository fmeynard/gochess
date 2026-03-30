# Codebase Overview

## Project purpose

This repository is a small Go chess engine prototype focused on legal move generation and perft-style validation.

The code is centered around:

- bitboard-backed position representation
- pseudo-legal move generation
- legal move filtering through `make` / `unmake`
- attack detection and check validation

There is no full search, evaluation, UCI loop, or user-facing game loop yet.

## Repository layout

- `cmd/perft.go`: current executable entry point; runs perft divide from a FEN and depth
- `cmd/main.go`: effectively unused placeholder
- `internal/position.go`: board state, FEN parsing, occupancy and piece queries
- `internal/bitsboard-move-generator.go`: pseudo-legal move generation and precomputed masks
- `internal/engine.go`: legal move generation and recursive perft
- `internal/position-updater.go`: `MakeMove`, `UnMakeMove`, and king-safety invalidation
- `internal/check.go`: attack detection and check evaluation
- `internal/move.go`: move representation
- `internal/move-history.go`: undo snapshot used by `UnMakeMove`
- `internal/tools.go`: square/index conversion and bit helpers
- `internal/*_test.go`: rule-specific tests

## Core design

### Position

`Position` is the main mutable board state.

It stores:

- active color
- castling rights for both sides
- en passant target square
- white and black king squares
- aggregate occupancy bitboards
- per-piece-type bitboards
- cached king safety state

The board is not stored as a `[64]Piece` array. Piece lookup is derived from occupancy plus piece-type bitboards.

### Move generator

`BitsBoardMoveGenerator` precomputes masks for:

- king attacks
- knight attacks
- pawn pushes and captures
- slider rays for bishops, rooks, and queens
- diagonal attack masks used by check detection

Pseudo-legal generation is split by piece family:

- `PawnPseudoLegalMoves`
- `KnightPseudoLegalMoves`
- `KingPseudoLegalMoves`
- `SliderPseudoLegalMoves`

Castling path emptiness is handled in king pseudo-legal generation.

### Engine

`Engine.LegalMoves` is the main move-generation entry point.

Flow:

1. Iterate active-side pieces from the occupancy bitboard.
2. Detect piece type from the per-piece bitboards.
3. Generate pseudo-legal targets.
4. Build a `Move`.
5. If the side is not currently in check and the move does not affect king lines, accept it early.
6. Otherwise `MakeMove`, test if the moving side king is still in check, then `UnMakeMove`.

This is a standard pseudo-legal plus legality-filtering design.

### Make / unmake

`PositionUpdater` mutates a shared `*Position`.

`MakeMove` currently handles:

- normal piece movement
- captures by overwriting destination square
- castling rook movement
- en passant capture resolution
- en passant target updates after double pawn moves
- castling-right updates when king or rook moves
- king position updates
- king-safety cache refresh when needed
- active color toggle

`UnMakeMove` restores state from `MoveHistory`, which is currently a near-full snapshot of the mutable position.

## Check detection

`check.go` implements attack tests for:

- pawns
- knights
- kings
- sliding pieces

Sliding attacks are split into:

- rank attacks
- file attacks
- diagonal attacks

`IsKingInCheck` uses these helpers and also updates the cached king-safety fields in `Position`.

## Current move / state model

`Move` stores:

- moved piece
- start square
- end square
- flag

Flags exist for:

- normal moves
- promotions
- en passant
- double pawn move
- castle
- capture

In practice, much of the current code infers move behavior from board state and squares rather than using the flag consistently.

## Current strengths

- The architecture is already suitable for perft and search-oriented evolution.
- Bitboard occupancy and per-piece boards are in place.
- `make` / `unmake` already exists.
- There is a decent set of tests for move generation, checks, en passant, castling, and undo behavior.
- The code is small enough to refactor safely.

## Current limitations and likely incomplete areas

### Promotions

Promotions are not fully integrated.

`PawnPseudoLegalMoves` returns a `promotionIdx`, but `Engine.LegalMoves` currently ignores it and does not emit promotion move variants.

### Move flags are underused

The move flag exists, but the codebase does not yet rely on it as the single source of truth for special move handling.

### Undo path is correct but heavy

`MoveHistory` stores most mutable position fields. This is simple and safe, but probably heavier than necessary for a hot perft/search path.

### Entrypoints are minimal

The only practical executable today is `cmd/perft.go`.

## Tests and execution notes

The repository includes tests for:

- position parsing
- piece helpers
- pseudo-legal move generation
- attack detection
- make/unmake behavior
- castling rights
- en passant updates

I was not able to run `go test ./...` in this environment because `go` is not installed on `PATH`.

## Short mental model

The current engine is:

- one mutable bitboard-based `Position`
- pseudo-legal generators per piece family
- legality filtering by `MakeMove` / `UnMakeMove`
- attack-based king safety checks
- perft-oriented validation tooling

If continuing work, the most natural next engine-level improvements are:

- complete promotion handling
- shrink `MoveHistory` into a compact undo record
- make move flags authoritative for special move behavior
- benchmark and reduce hot-path copying / allocations
