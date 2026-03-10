# Testing

## Standard Suite

```bash
go test ./...
```

Live CLI integration:

```bash
set -a
source ./secrets.env
set +a
go test ./internal/cli -run TestCLIIntegrationDeckFlow -count=1 -v
```

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
