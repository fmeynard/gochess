# gochess agent cache

- purpose: go chess engine. current center = legal movegen + make/unmake + attack detection + perft correctness/perf. not product-complete.
- current shipped surfaces: `cmd/perft.go`, `cmd/benchperft`, `cmd/uci`, `cmd/match`.
- current engine state is beyond placeholder stage:
  - `internal/search`: fixed-depth negamax alpha-beta, movetime iterative deepening, qsearch, TT, killer/history ordering, mate/stalemate terminal handling.
  - `internal/eval`: material, PST by phase, mobility, piece safety, king safety, passed pawns, simple pawn structure.
  - `internal/lichess`: reserved only; no real integration yet.

- package ownership:
  - `internal/board`: source of truth for mutable state. move representation/history. make/unmake. zobrist.
  - `internal/movegen`: pseudo-legal + legal generation, attack/check/pin analysis, magic bitboards.
  - `internal/engine`: orchestration/public API/perft glue.
  - dependency direction must stay `board <- movegen <- engine`.

- core mental model:
  - `board.Position` is authoritative state.
  - updater split:
    - `NewPlainPositionUpdater()`: mutation/undo only
    - `NewPositionUpdater()`: zobrist-decorated updater
  - legal generation center = `PseudoLegalMoveGenerator.LegalMovesInto(...)`
  - `engine.Engine` wires board + movegen + eval + search

- repo workflow:
  - normal flow = issue -> branch -> implement/validate -> PR -> merge
  - branch names: `issue-<n>-...` or focused topic names
  - every material change should go through PR
  - docs-only / repo-metadata changes do not need benchmarks

- validation rules:
  - baseline usually `go test ./...`
  - if touching `internal/board` or `internal/movegen`, also run:
    - `BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh`

- benchmark canonical context:
  - target = perft position 3
  - FEN = `8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1`
  - depth = 7
  - mode = hot
  - tricks = off
  - runner = compiled `benchperft`
  - use hot/no-tricks as primary perf signal; cold runs only for setup-cost questions
  - "perft tricks" means bulk counting at depth 2 + perft TT

- perf reference snapshot:
  - current best documented no-tricks/hot result: `v20`
  - nodes: `178,633,661`
  - time: `5.544416355s`
  - recorded reference shorthand: `5.54s`

- optimization heuristics worth remembering:
  - prefer changes that help future search too, not perft-only architecture
  - good ROI historically:
    - make/unmake hot-path specialization
    - attack/check analysis simplification
    - magic-bitboard slider lookups
    - eliminating large hot-path struct copies
    - fast-pathing common legality states
  - avoid retrying without new evidence:
    - broad bundled refactors across many perf axes
    - helper-heavy move-materialization splits that add branch/indirection cost
    - pointer-indirection replacement of simple board helper switches
    - compressed slider lookup indexing that makes lookup computation hot

- common commands:
  - tests: `go test ./...`
  - benchmark: `BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh`
  - direct divide: `go run ./cmd/perft.go '<fen>' 7`
  - build UCI: `mkdir -p ./bin && go build -o ./bin/gochess-uci ./cmd/uci`
  - smoke UCI: `printf 'uci\nisready\nposition startpos\ngo depth 1\nquit\n' | ./bin/gochess-uci`
  - local match smoke: `go run ./cmd/match -games 2 -movetime 50 -notes "sample self-play smoke run"`
  - default runner via make: `make runner`

- match runner defaults from docs:
  - `MODE=tui`
  - `CONCURRENT=5`
  - `OPPONENT_TAG=score-v1`
  - `GAMES=10`
  - `MOVETIME=1000`
  - `MOVE_OVERHEAD=50`
  - runner prints markdown row for manual paste into `docs/match-history.md`

- constraints/reminders:
  - keep docs updated when package layout, benchmark references, contributor workflow, or architecture assumptions change
  - tag scoring versions only after merge to `main`, then compare vs previous score tag
  - if session starts cold, read this file first, then only open deeper docs when the task needs specifics
