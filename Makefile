GOCACHE ?= $(CURDIR)/.codex-tmp/go-build-cache
MODE ?= tui
CONCURRENT ?= 5
OPPONENT_TAG ?= score-v1
GAMES ?= 10
MOVETIME ?= 1000
MOVE_OVERHEAD ?= 50
NOTES ?=
BENCH_DEPTH ?= 7
BENCH_MODE ?= hot
BENCH_WARMUP ?= 1
BENCH_NO_PERFT_TRICKS ?= 0
BENCH_FEN ?=
BENCH_PROFILE ?=
BOARD_BENCH ?= BenchmarkPositionUpdaterMakeUnmake
BOARD_BENCH_TIME ?= 200ms

RUNNER_FLAGS = -games $(GAMES) -parallel $(CONCURRENT) -movetime $(MOVETIME) -move-overhead $(MOVE_OVERHEAD)
ifneq ($(strip $(OPPONENT_TAG)),)
RUNNER_FLAGS += -opponent-tag $(OPPONENT_TAG)
endif
ifneq ($(strip $(NOTES)),)
RUNNER_FLAGS += -notes "$(NOTES)"
endif
ifeq ($(MODE),plain)
RUNNER_FLAGS += -plain
endif

BENCH_PERFT_ENV = GOCACHE="$(GOCACHE)" BENCH_DEPTH=$(BENCH_DEPTH) BENCH_MODE=$(BENCH_MODE) BENCH_WARMUP=$(BENCH_WARMUP) BENCH_NO_PERFT_TRICKS=$(BENCH_NO_PERFT_TRICKS)
ifneq ($(strip $(BENCH_FEN)),)
BENCH_PERFT_ENV += BENCH_FEN='$(BENCH_FEN)'
endif
ifneq ($(strip $(BENCH_PROFILE)),)
BENCH_PERFT_ENV += BENCH_PROFILE='$(BENCH_PROFILE)'
endif

.PHONY: help test build-uci smoke-uci runner bench-perft bench-perft-hot bench-perft-hot-no-tricks bench-perft-cold-no-tricks bench-updater-bench

help:
	@echo "Common targets:"
	@echo "  make test"
	@echo "  make build-uci"
	@echo "  make smoke-uci"
	@echo "  make runner MODE=tui CONCURRENT=5 OPPONENT_TAG=score-v1 GAMES=10 MOVETIME=1000 MOVE_OVERHEAD=50"
	@echo "  make bench-perft BENCH_DEPTH=7 BENCH_MODE=hot BENCH_NO_PERFT_TRICKS=1"
	@echo "  make bench-perft-hot"
	@echo "  make bench-perft-hot-no-tricks"
	@echo "  make bench-perft-cold-no-tricks"
	@echo "  make bench-updater-bench"
	@echo ""
	@echo "Runner variables:"
	@echo "  MODE=tui|plain        default: $(MODE)"
	@echo "  CONCURRENT=<n>        default: $(CONCURRENT)"
	@echo "  OPPONENT_TAG=<tag>    default: $(OPPONENT_TAG)"
	@echo "  GAMES=<n>             default: $(GAMES)"
	@echo "  MOVETIME=<ms>         default: $(MOVETIME)"
	@echo "  MOVE_OVERHEAD=<ms>    default: $(MOVE_OVERHEAD)"
	@echo "  NOTES=<text>          default: empty"
	@echo ""
	@echo "Benchmark variables:"
	@echo "  BENCH_DEPTH=<n>       default: $(BENCH_DEPTH)"
	@echo "  BENCH_MODE=hot|cold   default: $(BENCH_MODE)"
	@echo "  BENCH_WARMUP=0|1      default: $(BENCH_WARMUP)"
	@echo "  BENCH_NO_PERFT_TRICKS=0|1  default: $(BENCH_NO_PERFT_TRICKS)"
	@echo "  BENCH_FEN='<fen>'     default: benchmark script default"
	@echo "  BENCH_PROFILE=<path>  default: benchmark script default"
	@echo "  BOARD_BENCH=<regex>   default: $(BOARD_BENCH)"
	@echo "  BOARD_BENCH_TIME=<d>  default: $(BOARD_BENCH_TIME)"

test:
	GOCACHE="$(GOCACHE)" go test ./...

build-uci:
	mkdir -p ./bin
	GOCACHE="$(GOCACHE)" go build -o ./bin/gochess-uci ./cmd/uci

smoke-uci: build-uci
	printf 'uci\nisready\nposition startpos\ngo depth 1\nquit\n' | ./bin/gochess-uci

runner:
	GOCACHE="$(GOCACHE)" go run ./cmd/match $(RUNNER_FLAGS)

bench-perft:
	$(BENCH_PERFT_ENV) ./scripts/bench-perft.sh

bench-perft-hot:
	$(MAKE) bench-perft BENCH_MODE=hot BENCH_WARMUP=1

bench-perft-hot-no-tricks:
	$(MAKE) bench-perft BENCH_MODE=hot BENCH_WARMUP=1 BENCH_NO_PERFT_TRICKS=1

bench-perft-cold-no-tricks:
	$(MAKE) bench-perft BENCH_MODE=cold BENCH_WARMUP=0 BENCH_NO_PERFT_TRICKS=1

bench-updater-bench:
	GOCACHE="$(GOCACHE)" go test ./internal/board -run '^$$' -bench '$(BOARD_BENCH)' -benchtime=$(BOARD_BENCH_TIME)
