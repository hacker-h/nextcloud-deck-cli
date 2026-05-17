# Testing

## Standard Suite

```bash
go test ./...
```

Build the CLI locally:

```bash
go build ./cmd/deck
```

Timeout-sensitive slow runs can use the global timeout flag or environment variable:

```bash
DECK_TIMEOUT=10m go test ./internal/deck -run TestPerformanceBoardBackupImport -count=1 -v
deck --timeout 10m board import --file ./large-board.json
```

Live CLI integration:

```bash
set -a
source ./secrets.env
set +a
go test ./internal/cli -run TestCLIIntegrationDeckFlow -count=1 -v
```

`secrets.env` is intentionally untracked. It should define:

```bash
NEXTCLOUD_BASE_URL=https://nextcloud.example.com
NEXTCLOUD_USERNAME=example-user
NEXTCLOUD_APP_PASSWORD=example-app-password
DECK_TIMEOUT=10m
```

The low-cost live smoke checks are:

```bash
go test ./internal/deck -run TestIntegrationGetBoards -count=1 -v
go test ./internal/cli -run TestCLIIntegrationDeckFlow -count=1 -timeout 20m -v
```

## CI Live Integration

The `Live Integration` GitHub Actions workflow runs on pushes to `main`, on a weekly schedule, and manually through `workflow_dispatch`. It intentionally does not run on pull requests, so public PRs do not receive test-server secrets. Configure these repository secrets before running it:

- `NEXTCLOUD_BASE_URL`
- `NEXTCLOUD_USERNAME`
- `NEXTCLOUD_APP_PASSWORD` or `NEXTCLOUD_PASSWORD`

The workflow always builds the CLI, runs the live client smoke test, and runs the broad CLI feature flow. Server-version or permission-limited commands are logged and tracked in GitHub issues instead of blocking unrelated feature coverage. Optional manual inputs enable the slower performance/import-export and rich backup/restore scenarios.

## Performance

```bash
set -a
source ./secrets.env
set +a
go test ./internal/deck -run TestPerformanceLargeBoard -count=1 -v
go test ./internal/deck -run TestPerformanceBoardBackupImport -count=1 -v
```

Measured results live in `BENCHMARKS.md`.

## Full Backup / Restore Scenario

Heavy end-to-end restore scenario:

```bash
set -a
source ./secrets.env
export NEXTCLOUD_FULL_BACKUP_SCENARIO=1
set +a
go test ./internal/deck -run TestBackupRestoreRichKanbanBoard -count=1 -v
```

What it builds and verifies:

- a 6-lane kanban board
- 200 cards total
- realistic workflow states across `Inbox`, `Backlog`, `Ready`, `Doing`, `Blocked`, `Done`
- representative cards with due dates, labels, comments, attachments, and markdown todo lists
- cards that were moved between lanes before backup
- restore assertions for lane counts, total card count, moved-card placement, checklist preservation, and due-date preservation

This is intentionally opt-in because it is expensive and only makes sense once full backup/restore is trustworthy enough for your server.
