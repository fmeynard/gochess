# Future Optimisations

This document tracks the remaining move-generation performance ideas after the `v8` work on move flags, undo-record compaction, cached king-affect restoration, and optional Zobrist maintenance.

## 1. Occupancy-Driven Slider Attacks

The current slider move generation still walks precomputed square lists one square at a time.

Relevant code:

- `internal/pseudo-legal-move-generator.go`
- `SliderPseudoLegalMovesInto(...)`

Likely next step:

- Replace directional square scans with occupancy-driven attack generation
- Candidate techniques: magic bitboards, PEXT-based lookups, or hyperbola quintessence

Expected benefit:

- Fewer branches in rook/bishop/queen move generation
- Better scaling on dense middlegame positions
- Cleaner reuse of attack masks for both move generation and attack detection

## 2. Bitboard-First Legal Filtering

`legalMovesInto(...)` is already much better than the earlier make/unmake-heavy path, but it still works as:

1. Generate target squares into a temporary array
2. Apply pin / evasion / king-safety logic per target
3. Materialize `Move` values

Relevant code:

- `internal/engine.go`
- `legalMovesInto(...)`

Likely next step:

- Generate destination bitboards first
- Mask them with:
  - `^friendlyOcc`
  - pin rays
  - evasion masks
  - king-safe destination masks
- Iterate only the surviving bits

Expected benefit:

- Less branching in the inner loop
- Better composition of legality constraints
- A cleaner path toward specialized legal generators per piece type

## 3. Pawn Generation Specialization

Pawn generation is still one of the hotter pseudo-legal generators because it mixes:

- side-to-move branching
- promotions
- double pushes
- en passant

Relevant code:

- `internal/pseudo-legal-move-generator.go`
- `PawnPseudoLegalMovesInto(...)`

Likely next step:

- Split white and black pawn generation into separate specialized functions
- Remove repeated color-dependent branching from the hot path
- Consider emitting captures and pushes from bitboard expressions instead of mixed conditional logic

Expected benefit:

- Lower branch pressure
- Simpler hot code per side

## 4. More Specialized Board Updates

`movePiece(...)` and `capturePiece(...)` are already much better than generic board rewriting, but they still switch on piece type on every move.

Relevant code:

- `internal/position.go`
- `movePiece(...)`
- `capturePiece(...)`

Likely next step:

- Add specialized fast paths for the most common cases:
  - pawn quiet move
  - pawn capture
  - king move
  - rook move
- Keep the generic helpers as fallback

Expected benefit:

- Lower dispatch overhead in `MakeMove` / `UnMakeMove`
- Better inlining opportunities

## 5. Even Tighter Move Representation

`Move` already carries more information than before, but the representation is still structurally “field-based”.

Relevant code:

- `internal/move.go`

Likely next step:

- Pack move data into a narrower integer representation
- Keep cheap accessors for:
  - start square
  - end square
  - moving piece
  - move kind

Expected benefit:

- Reduced copy cost for move buffers and undo records
- Better cache density in deep perft/search trees

## 6. Profile Search Separately From Perft

The current benchmark work is strongly perft-driven. That is useful, but some optimizations may help perft more than search, or the reverse.

Likely next step:

- Add a repeatable search benchmark suite
- Track move-generation, make/unmake, and attack-detection costs under search workloads

Expected benefit:

- Avoid overfitting the engine purely to perft
- Make tradeoffs more explicit before larger architectural work

## Suggested Order

If the goal remains “highest performance gain for the lowest implementation risk”, the next sequence to test is:

1. Specialized pawn generation
2. More specialized board-update helpers
3. Bitboard-first legal filtering
4. Occupancy-driven slider attack generation
5. Tighter move packing
