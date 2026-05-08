# nextcloud-deck-api

Go CLI for Nextcloud Deck.

## Get Started

Env:

- `NEXTCLOUD_BASE_URL`
- `NEXTCLOUD_USERNAME`
- `NEXTCLOUD_PASSWORD` or `NEXTCLOUD_APP_PASSWORD`

Example:

```bash
export NEXTCLOUD_BASE_URL="https://nextcloud.example.com"
export NEXTCLOUD_USERNAME="antonia"
export NEXTCLOUD_PASSWORD="secret"
```

Build:

```bash
go build ./cmd/deck
```

Quick start:

```bash
deck board create --title "Project" --color ff6600
deck board find --title "Project"
deck list create --board 1 --title "Backlog"
deck list find --board 1 --title "Backlog"
deck label find --board 1 --title "Bug"
deck card create --board 1 --stack 2 --title "Test"
deck card describe --board 1 --stack 2 --card 3 --description "- [ ] follow up"
deck todo add --board 1 --stack 2 --card 3 --text "Call customer"
deck board export --board 1 --out ./board.json
deck board import --file ./board.json
```

## Output

Commands default to plain text. Use `--json`, `-o json`, or `--output json` for machine-readable JSON output.

```bash
deck board list
deck board list --json
deck card move --board 1 --from-stack 2 --to-stack 3 --card 4 -o json
```

Supported output formats:

- `text` default
- `json`

Boolean aliases are also accepted: `--json` and `--text`.

## Commands

`find` commands use exact, case-sensitive title matches and fail when a title is missing or duplicated.

`board`

```bash
deck board list [--details]
deck board get --board ID
deck board find --title TEXT
deck board create --title TEXT [--color HEX]
deck board update --board ID [--title TEXT] [--color HEX]
deck board archive --board ID
deck board unarchive --board ID
deck board clone --board ID [--with-cards BOOL] [--with-assignments BOOL] [--with-labels BOOL] [--with-due-date BOOL] [--move-cards-left BOOL] [--restore-archived-cards BOOL]
deck board export --board ID --out PATH
deck board import --file PATH
deck board import-systems
deck board import-schema --name NAME
deck board delete --board ID
deck board restore --board ID
```

`list`

```bash
deck list list --board ID
deck list archived --board ID
deck list get --board ID --list ID
deck list find --board ID --title TEXT
deck list create --board ID --title TEXT [--order N]
deck list rename --board ID --list ID --title TEXT
deck list reorder --board ID --list ID --order N
deck list delete --board ID --list ID
```

`card`

```bash
deck card list --board ID --stack ID
deck card get --board ID --stack ID --card ID
deck card create --board ID --stack ID --title TEXT [--description TEXT] [--due RFC3339] [--order N]
deck card clone --card ID --to-stack ID
deck card rename --board ID --stack ID --card ID --title TEXT
deck card describe --board ID --stack ID --card ID --description TEXT
deck card move --board ID --from-stack ID --to-stack ID --card ID [--order N]
deck card reorder --board ID --stack ID --card ID --order N
deck card archive --board ID --stack ID --card ID
deck card unarchive --board ID --stack ID --card ID
deck card done --card ID
deck card undone --card ID
deck card due get --board ID --stack ID --card ID
deck card due set --board ID --stack ID --card ID --value RFC3339
deck card due clear --board ID --stack ID --card ID
deck card assign-user --board ID --stack ID --card ID --user USER
deck card unassign-user --board ID --stack ID --card ID --user USER
deck card assign-label --board ID --stack ID --card ID --label ID
deck card remove-label --board ID --stack ID --card ID --label ID
deck card delete --board ID --stack ID --card ID
```

`todo`

```bash
deck todo list --board ID --stack ID --card ID
deck todo add --board ID --stack ID --card ID --text TEXT
deck todo check --board ID --stack ID --card ID --index N
deck todo uncheck --board ID --stack ID --card ID --index N
```

`label`

```bash
deck label list --board ID
deck label get --board ID --label ID
deck label find --board ID --title TEXT
deck label create --board ID --title TEXT [--color HEX]
deck label update --board ID --label ID [--title TEXT] [--color HEX]
deck label delete --board ID --label ID
```

`comment`

```bash
deck comment list --card ID
deck comment create --card ID --message TEXT
deck comment update --card ID --comment ID --message TEXT
deck comment delete --card ID --comment ID
```

`attachment`

```bash
deck attachment list --board ID --stack ID --card ID
deck attachment upload --board ID --stack ID --card ID --file PATH
deck attachment download --board ID --stack ID --card ID --attachment ID --out PATH
deck attachment delete --board ID --stack ID --card ID --attachment ID
deck attachment restore --board ID --stack ID --card ID --attachment ID
```

`share`

```bash
deck share list --board ID
deck share create --board ID --type N --participant VALUE [--edit BOOL] [--share BOOL] [--manage BOOL]
deck share update --board ID --share-id ID [--edit BOOL] [--share BOOL] [--manage BOOL]
deck share delete --board ID --share-id ID
```

`config`

```bash
deck config get
deck config set --key KEY --value VALUE
```

`search`, `overview`, `user`, `capabilities`, `activity`

```bash
deck search cards --term TEXT [--limit N]
deck overview upcoming
deck user search --term TEXT
deck user get --user USER
deck capabilities
deck activity card --card ID
```

`session`

```bash
deck session create --board ID
deck session sync --board ID --token TOKEN
deck session close --board ID --token TOKEN
```

## Testing

```bash
go test ./...
```

Live integration:

```bash
set -a
source ./secrets.env
set +a
go test ./internal/cli -run TestCLIIntegrationDeckFlow -count=1 -v
```

Performance and benchmark results:

- `BENCHMARKS.md`

## Not Implemented Yet

- `board import-systems` and `board import-schema` are exposed but not treated as implemented on the verified server because it returns `404`
- `board restore` is exposed but not treated as implemented on the verified server because it returns `403`
- `session` commands are exposed but not treated as fully reliable on the verified server
- todos are markdown checkboxes in descriptions, not a native structured checklist API
