#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

mkdir -p .codex-tmp

fen="${BENCH_FEN:-8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1}"
depth="${BENCH_DEPTH:-6}"
profile="${BENCH_PROFILE:-.codex-tmp/bench-perft.cpu.prof}"
mode="${BENCH_MODE:-hot}"
warmup="${BENCH_WARMUP:-1}"
extra_args=()

if [[ "${BENCH_NO_PERFT_TRICKS:-0}" == "1" ]]; then
  extra_args+=("-no-perft-tricks")
fi

GOCACHE="${GOCACHE:-$repo_root/.codex-tmp/go-build-cache}" \
  go build -o .codex-tmp/benchperft-bin ./cmd/benchperft

./.codex-tmp/benchperft-bin \
  -fen "$fen" \
  -depth "$depth" \
  -mode "$mode" \
  -warmup "$warmup" \
  -cpuprofile "$profile" \
  "${extra_args[@]}"
