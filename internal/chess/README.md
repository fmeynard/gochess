# Chess Package

`internal/chess` currently contains the full chess domain package:

- board state
- move representation
- pseudo-legal and legal move generation
- attack detection and position analysis
- make/unmake
- zobrist hashing
- perft orchestration

This package boundary is intentional for now.

Future work is expected to extract higher-level packages around it rather than immediately splitting this package along leaky internal seams:

- `internal/search`
- `internal/eval`
- `internal/lichess`

Once the internal APIs are cleaner, move generation may be extracted further without forcing large-scale symbol export churn first.
