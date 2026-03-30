# Future Optimisations

## MoveHistory and undo strategy

The current `MoveHistory` approach is correct in spirit because a chess engine should keep an undo record and use `make` / `unmake` in the hot path.

What is probably not optimal is storing a near-full snapshot of the mutable position for every move:

- `occupied`
- `whiteOccupied`
- `blackOccupied`
- all piece bitboards
- king state caches
- castling rights
- en passant square

This makes `UnMakeMove` simple, but it copies more data than necessary at each node.

## Better direction

Prefer a compact incremental undo record instead of:

- recomputing the previous position from the last move
- storing a full position snapshot

The engine should:

1. Apply the move incrementally.
2. Save only the state that cannot be reconstructed cheaply.
3. Reverse the exact delta during `UnMakeMove`.

## Why not recompute from the previous move

Recomputing the old position from the last move is usually slower and more fragile because undo still needs enough information to restore:

- the captured piece
- previous castling rights
- previous en passant square
- king locations
- promotion state
- en passant capture square
- rook movement during castling

If that information must already be stored, then a compact undo record is usually the better design.

## Suggested compact undo record

A lean undo structure for this engine could store:

- moved piece
- captured piece
- start square
- end square
- move flag / move kind
- previous white castling rights
- previous black castling rights
- previous en passant square
- previous white king safety cache
- previous black king safety cache
- previous white king square if needed
- previous black king square if needed

Depending on the final `Move` representation, some of these fields may already exist on the move itself and do not need to be duplicated in the undo record.

## Incremental unmake logic

With the captured piece and move kind available, the engine can restore bitboards directly instead of restoring full snapshots.

Typical `UnMakeMove` flow:

1. Remove the moved piece from the destination square.
2. Restore the moved piece on the origin square.
3. If the move was a capture, restore the captured piece on the correct square.
4. If the move was en passant, restore the pawn on the passed square instead of the destination square.
5. If the move was castling, move the rook back.
6. If the move was a promotion, remove the promoted piece and restore a pawn on the origin square.
7. Restore castling rights, en passant square, active color, and cached king safety values.

## Bitboards that can be restored incrementally

The following do not need to be copied wholesale if undo data is sufficient:

- `occupied`
- `whiteOccupied`
- `blackOccupied`
- `pawnBoard`
- `knightBoard`
- `bishopBoard`
- `rookBoard`
- `queenBoard`
- `kingBoard`

They can be updated by applying and reversing the move delta.

## Practical guidance

- Keep passing `*Position` in the hot path.
- Keep `make` / `unmake`.
- Replace full `MoveHistory` snapshots with a compact undo record.
- Store the captured piece explicitly.
- Ensure move flags are rich enough to distinguish normal moves, captures, castling, en passant, and promotions.

## Expected benefit

This should reduce per-node memory traffic and improve cache behavior while keeping undo constant-time and deterministic.

The expected performance win is more likely to come from reducing repeated state copying than from changing `Position` from pointer to value.

## Memory and GC

If GC time grows quickly at deeper perft depths, the likely cause is not recursion itself but heap allocation inside the recursive path.

In a move tree, even small per-node allocations become expensive because they are multiplied by every node in the tree.

### Likely allocation sources in the current code

- `Engine.LegalMoves` creates a fresh `[]Move` on each call.
- Pseudo-legal generators create fresh `[]int8` slices on each call.
- `MakeMove` creates a new `MoveHistory` for each move.
- The current `MoveHistory` copies a large amount of mutable state.

### Important principle

Recursion is not automatically the GC problem.

The usual problem is:

- recursive traversal
- plus per-node heap allocations
- multiplied by millions of nodes

The goal for perft and search should be to make the hot path close to allocation-free.

### Recommended direction

- Keep recursion if it stays simple and fast.
- Remove heap allocations from the recursive move-generation path.
- Use compact undo data instead of large snapshots.
- Prefer preallocated per-ply buffers for moves.
- Prefer writing moves directly into a caller-owned buffer.

### Concrete opportunities

#### 1. Generate moves directly into a buffer

Avoid generating pseudo-legal `[]int8` slices only to immediately convert each target square into a `Move`.

Instead:

- pass a destination buffer into move generation
- append `Move` values directly
- return only the count written

This removes an intermediate allocation layer and reduces temporary objects in the hottest path.

#### 2. Preallocate move storage per ply

For perft/search paths, use a fixed per-depth move buffer such as:

- `[MaxPly][MaxMoves]Move`

Then each depth level writes into its own slice window.

This avoids allocating new move slices at each recursive call.

#### 3. Replace pointer undo records with compact value undo records

If possible:

- return undo data by value
- keep the undo struct small
- avoid copying all bitboards every move

#### 4. Benchmark allocation counts directly

When Go is available, use:

- `go test -bench . -benchmem`
- `go build -gcflags=-m ./...`

The target should be to drive allocations in the perft path toward zero or near-zero.
