# nextcloud-deck-api

Go CLI for Nextcloud Deck with a typed client, command-line workflows, and real integration tests against a live Nextcloud instance.

## What It Does

- Manages boards, lists, cards, labels, comments, attachments, shares, sessions, config, search, upcoming cards, capabilities, users, and card activity
- Supports markdown-based card todos through checklist parsing in card descriptions
- Works as both a reusable Go client and a CLI
- Includes end-to-end integration coverage against a real Deck server

## Environment

Required:

- `NEXTCLOUD_BASE_URL`
- `NEXTCLOUD_USERNAME`
- `NEXTCLOUD_PASSWORD` or `NEXTCLOUD_APP_PASSWORD`

Example:

```bash
export NEXTCLOUD_BASE_URL="https://nextcloud.example.com"
export NEXTCLOUD_USERNAME="antonia"
export NEXTCLOUD_PASSWORD="secret"
```

The loader also accepts host-only values like `nextcloud.example.com` and normalizes them to `https://...`.

## Build

```bash
go build ./cmd/deck
```

Binary path if built in-place:

```bash
./deck
```

Or run directly:

```bash
go run ./cmd/deck --help
```

## Quick Start

```bash
deck board create --title "Project" --color ff6600
deck list create --board 1 --title "Backlog"
deck card create --board 1 --stack 2 --title "Test"
deck card describe --board 1 --stack 2 --card 3 --description "- [ ] follow up"
deck todo add --board 1 --stack 2 --card 3 --text "Call customer"
deck card move --board 1 --from-stack 2 --to-stack 4 --card 3 --order 999
deck board export --board 1 --out ./board.json
deck board import --file ./board.json
```

## Command Reference

### Board Commands

List boards:

```bash
deck board list
deck board list --details --json
```

Get one board:

```bash
deck board get --board 1
deck board get --board 1 --json
```

Create a board:

```bash
deck board create --title "Project" --color ff6600
```

Update a board:

```bash
deck board update --board 1 --title "Project Alpha" --color 00aa88
```

Archive or unarchive a board:

```bash
deck board archive --board 1
deck board unarchive --board 1
```

Clone a board:

```bash
deck board clone --board 1 --with-cards true --with-labels true --with-due-date true
deck board clone --board 1 --with-cards true --with-assignments true --move-cards-left true
```

Export or import a board:

```bash
deck board export --board 1 --out ./board.json
deck board import --file ./board.json
```

List import systems or inspect an import schema:

```bash
deck board import-systems
deck board import-schema --name DeckJson
```

Delete or restore a board:

```bash
deck board delete --board 1
deck board restore --board 1
```

### List Commands

List active or archived lists:

```bash
deck list list --board 1
deck list archived --board 1
```

Get one list:

```bash
deck list get --board 1 --list 4
```

Create a list:

```bash
deck list create --board 1 --title "Backlog" --order 10
```

Rename or reorder a list:

```bash
deck list rename --board 1 --list 4 --title "Doing"
deck list reorder --board 1 --list 4 --order 20
```

Delete a list:

```bash
deck list delete --board 1 --list 4
```

### Card Commands

List or get cards:

```bash
deck card list --board 1 --stack 2
deck card get --board 1 --stack 2 --card 3
```

Create a card:

```bash
deck card create --board 1 --stack 2 --title "Test"
deck card create --board 1 --stack 2 --title "Test" --description "Draft" --due "2026-03-08T12:00:00Z"
```

Clone a card into another list:

```bash
deck card clone --card 3 --to-stack 5
```

Rename or describe a card:

```bash
deck card rename --board 1 --stack 2 --card 3 --title "Renamed"
deck card describe --board 1 --stack 2 --card 3 --description "- [ ] todo"
```

Move or reorder a card:

```bash
deck card move --board 1 --from-stack 2 --to-stack 5 --card 3 --order 999
deck card reorder --board 1 --stack 5 --card 3 --order 10
```

Archive, unarchive, complete, or undo completion:

```bash
deck card archive --board 1 --stack 5 --card 3
deck card unarchive --board 1 --stack 5 --card 3
deck card done --card 3
deck card undone --card 3
```

Delete a card:

```bash
deck card delete --board 1 --stack 5 --card 3
```

Due date workflows:

```bash
deck card due get --board 1 --stack 2 --card 3
deck card due set --board 1 --stack 2 --card 3 --value "2026-03-08T12:00:00Z"
deck card due clear --board 1 --stack 2 --card 3
```

User and label assignment:

```bash
deck card assign-user --board 1 --stack 2 --card 3 --user antonia
deck card unassign-user --board 1 --stack 2 --card 3 --user antonia
deck card assign-label --board 1 --stack 2 --card 3 --label 9
deck card remove-label --board 1 --stack 2 --card 3 --label 9
```

### Todo Commands

Todos are stored as markdown checkboxes inside the card description.

List todos on a card:

```bash
deck todo list --board 1 --stack 2 --card 3
```

Add a todo:

```bash
deck todo add --board 1 --stack 2 --card 3 --text "Call customer"
```

Check or uncheck a todo:

```bash
deck todo check --board 1 --stack 2 --card 3 --index 1
deck todo uncheck --board 1 --stack 2 --card 3 --index 1
```

### Label Commands

```bash
deck label list --board 1
deck label get --board 1 --label 9
deck label create --board 1 --title "Blocked" --color FF7A66
deck label update --board 1 --label 9 --title "Ready" --color 31CC7C
deck label delete --board 1 --label 9
```

### Comment Commands

```bash
deck comment list --card 3
deck comment create --card 3 --message "Need review"
deck comment update --card 3 --comment 7 --message "Reviewed"
deck comment delete --card 3 --comment 7
```

### Attachment Commands

```bash
deck attachment list --board 1 --stack 2 --card 3
deck attachment upload --board 1 --stack 2 --card 3 --file ./notes.txt
deck attachment download --board 1 --stack 2 --card 3 --attachment 12 --out ./notes-copy.txt
deck attachment delete --board 1 --stack 2 --card 3 --attachment 12
deck attachment restore --board 1 --stack 2 --card 3 --attachment 12
```

### Share Commands

Participant type values match Deck ACL types:

- `0` user
- `1` group
- `7` circle

```bash
deck share list --board 1
deck share create --board 1 --type 0 --participant antonia --edit true --share false --manage false
deck share update --board 1 --share-id 4 --edit true --share true --manage false
deck share delete --board 1 --share-id 4
```

### Config Commands

```bash
deck config get
deck config set --key cardIdBadge --value true
deck config set --key calendar --value false
```

### Search, Overview, Session, User, Capability, and Activity Commands

Search cards:

```bash
deck search cards --term "invoice" --limit 10
```

Upcoming cards:

```bash
deck overview upcoming
```

Session lifecycle:

```bash
deck session create --board 1
deck session sync --board 1 --token TOKEN
deck session close --board 1 --token TOKEN
```

User search and lookup:

```bash
deck user search --term anton
deck user get --user antonia
```

Capabilities and card activity:

```bash
deck capabilities
deck activity card --card 3
```

## Output

- Most write commands return JSON objects
- Some simple delete and session commands print a short status line
- `board list` defaults to tabular output; use `--json` for structured results

## Testing

Run unit and package tests:

```bash
go test ./...
```

Run the live Deck integration flow after sourcing credentials:

```bash
set -a
source ./secrets.env
set +a
go test ./internal/cli -run TestCLIIntegrationDeckFlow -count=1 -v
```

The integration suite creates a temporary board, lists, cards, labels, comments, attachments, exports a board, imports it again, and cleans up as much as the server allows.

## Notes

- Some Nextcloud Deck API docs are stale; this project uses the public API where it works and the app routes where newer Deck UI behavior requires them
- `board import-systems` and `board import-schema` exist in code, but your current test server returns `404` for them
- `board restore` exists in code, but your current test account/server returns `403`
- Todo support is implemented via markdown checkboxes because Deck does not expose a dedicated public checklist API
