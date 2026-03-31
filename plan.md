# Plan

Primary short-term target:

- Get `perft(7)` on `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1` below `1s`
- Mode: `BENCH_NO_PERFT_TRICKS=1`

Constraints:

- Treat `perft` as a profiling and correctness tool for move generation, not as the long-term design center.
- Prefer optimizations that remain useful for a future search/scoring engine.
- Avoid perft-only architectural contortions unless they clearly dominate and are isolated.
- Use the `hot` benchmark mode as the default reference so setup time does not pollute movegen measurements.

Current best:

- `benchmark-v14`
- `perft(7)`: `6.857738422s`
- Nodes: `178,633,661`

Near-term work order:

1. Move generation hot paths
   - Optimize `appendKingMoves(...)`
   - Optimize `appendPawnMoves(...)`
   - Reduce `appendMovesFromMask(...)` overhead without reintroducing the failed helper split

2. Attack and legality pipeline
   - Re-profile `computePositionAnalysis(...)`
   - Re-profile `isSquareAttacked(...)`
   - Look for shared attack computations or cheaper legal filtering

3. Sliders
   - Evaluate denser rook/bishop attack lookup strategies
   - Consider magic-style indexing if the code/complexity tradeoff is favorable

4. Updater follow-up
   - Use updater microbenchmarks before changing `MakeMove(...)` / `UnMakeMove(...)`
   - Keep specializing only when a narrow path measures better

5. Search-ready architecture
   - Preserve a clean engine core for later scoring/search/UCI work
   - Keep move generation, attack logic, and make/unmake semantics trustworthy

Process for each optimization slice:

1. Form a profiling hypothesis
2. Make a narrow change
3. Run tests
4. Run `BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh`
5. Keep only changes that improve the benchmark and preserve code quality

Context files:

- `benchmark-history.md`
- `benchmark-learnings.md`
- `codebase-overview.md`
