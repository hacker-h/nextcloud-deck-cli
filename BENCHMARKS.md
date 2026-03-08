# Benchmarks

Live benchmark measurements collected against the dedicated Nextcloud test instance on 2026-03-08.

Commands used:

```bash
set -a
source ./secrets.env
set +a

go test ./internal/deck -run TestPerformanceLargeBoard -count=1 -v
go test ./internal/deck -run TestPerformanceBoardBackupImport -count=1 -v
```

## Large Board Operations

Fixture shape:

- 1 board
- 2 stacks
- 200 total cards
- move test moves 50 cards from stack A to stack B

Measured results:

| Operation | Result |
| --- | ---: |
| Create 100 cards sequentially | 1m33.321s |
| Average sequential create | 933ms/card |
| Create 100 cards in parallel (8 workers) | 18.845s |
| Average parallel create | 188ms/card |
| Parallel speedup vs sequential | 4.95x |
| Fetch stack with 200 cards | 657ms |
| Fetch board details with 2 stacks | 585ms |
| Move 50 cards sequentially | 2m25.052s |
| Average move time | 2.90s/card |
| Fetch both stacks after moves | 1.305s |
| Search cards on the large board | 601ms |

Takeaways:

- Card creation benefits strongly from bounded client-side parallelism.
- Read-heavy operations on 200-card boards stay well under 1.5 seconds in this environment.
- Card move/reorder is by far the slowest workflow and appears to be the main server-side bottleneck.

## Board Export / Backup / Import

Fixture shape:

- 1 board
- 2 stacks
- cards spread across both stacks
- each card includes a short description and markdown checklist items

Measured results:

| Board Size | Export Time | Export Size | Size Per Card | Import Time |
| --- | ---: | ---: | ---: | ---: |
| 10 cards | 587ms | 7,202 bytes | 720 bytes/card | 12.994s |
| 100 cards | 811ms | 61,665 bytes | 617 bytes/card | 1m38.050s |

Import verification:

- 10-card import recreated 2 stacks and 10 cards
- 100-card import recreated 2 stacks and 100 cards

Takeaways:

- Board export is fast and scales far better than card move operations.
- Export cost is close to flat between 10 and 100 cards in this setup.
- Import is much slower because the current server-side import path fails and the CLI falls back to client-side reconstruction.

## Deep Run Notes

- A deeper 200-card export/import run is intentionally disabled by default.
- Enable it with `NEXTCLOUD_PERF_DEEP=1` when you want heavier measurements.
- The deep import path is expensive enough that it can exceed normal `go test` timeout expectations.
