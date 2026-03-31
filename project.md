# Project Usage

## Build The UCI Engine

Build the UCI entrypoint:

```bash
mkdir -p ./bin
go build -o ./bin/gochess-uci ./cmd/uci
```

## Quick Local Smoke Test

You can talk to the engine directly:

```bash
printf 'uci\nisready\nposition startpos\ngo depth 1\nquit\n' | ./bin/gochess-uci
```

Expected protocol shape:

- `id name GoChess`
- `uciok`
- `readyok`
- one `info ...` line
- one `bestmove ...` line

## Load In A Standard GUI

The engine speaks UCI, so it should load in standard UCI GUIs such as Arena, BanksiaGUI, Cute Chess GUI, or similar tools.

Generic setup flow:

1. Build `./bin/gochess-uci`
2. Open your chess GUI
3. Add a new engine
4. Choose `UCI` as the engine protocol
5. Point the GUI to the binary:
   `.../gochess/bin/gochess-uci`
6. Start a test game or analysis session

## Supported Commands Today

Current UCI support includes:

- `uci`
- `isready`
- `ucinewgame`
- `position startpos ...`
- `position fen ...`
- `go depth N`
- `go movetime N`
- `stop`
- `quit`

Current notes:

- `stop` interrupts an in-flight search and returns the best move from the last completed work available
- advanced UCI options are not implemented yet
- the engine is already usable in a GUI, but the protocol surface will continue to improve

## Run Local Matches

Build and run a local self-play smoke match:

```bash
go run ./cmd/match -games 2 -movetime 50 -notes "sample self-play smoke run"
```

Run the current engine against a tagged revision that already contains `cmd/uci`:

```bash
go run ./cmd/match -opponent-tag <tag> -games 4 -movetime 5000
```

The runner prints a markdown row you can copy manually into `docs/match-history.md`.
