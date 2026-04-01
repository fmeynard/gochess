# GoChess

Go chess engine focused on:

- legal move generation
- `make` / `unmake`
- attack detection
- perft correctness and performance work

The project is not a full playing engine yet. Search, evaluation, self-play orchestration, and Lichess integration are planned but not implemented.

## Layout

- `internal/board`
  Mutable board state, move representation, make/unmake, Zobrist hashing.
- `internal/movegen`
  Pseudo-legal move generation, legal filtering, checks, pins, and attack analysis.
- `internal/engine`
  Engine orchestration and perft recursion.
- `cmd/perft.go`
  Simple perft entrypoint.
- `cmd/benchperft/main.go`
  Benchmark entrypoint used by `scripts/bench-perft.sh`.
- `cmd/uci/main.go`
  UCI entrypoint for GUI integration and external engine tooling.
- `docs/`
  Architecture notes, benchmark history, and optimization notes.

## Common Commands

Run tests:

```bash
go test ./...
```

Run the reference perft benchmark:

```bash
BENCH_DEPTH=7 BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1 ./scripts/bench-perft.sh
```

Run a direct perft divide:

```bash
go run ./cmd/perft.go '8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1' 7
```

Build the UCI binary:

```bash
mkdir -p ./bin
go build -o ./bin/gochess-uci ./cmd/uci
```

## Documentation

- [Contributing](./CONTRIBUTING.md)
- [Project Usage](./project.md)
- [Docs Index](./docs/README.md)
- [Codebase Overview](./docs/codebase-overview.md)
- [Benchmark History](./docs/benchmark-history.md)
- [Benchmark Learnings](./docs/benchmark-learnings.md)

## Useful External Links

- [FEN Position generator](http://www.netreal.de/Forsyth-Edwards-Notation/index.php?)
- [Chess FEN Viewer](https://www.dailychess.com/chess/chess-fen-viewer.php)
