# nextcloud-deck-api

Go CLI for Nextcloud Deck.

Env:

- `NEXTCLOUD_BASE_URL`
- `NEXTCLOUD_USERNAME`
- `NEXTCLOUD_PASSWORD` or `NEXTCLOUD_APP_PASSWORD`

Build:

```bash
go build ./cmd/deck
```

Examples:

```bash
deck board list --details --json
deck board create --title "Project"
deck list list --board 1
deck card create --board 1 --stack 2 --title "Test"
deck card rename --board 1 --stack 2 --card 3 --title "Renamed"
deck card describe --board 1 --stack 2 --card 3 --description "- [ ] todo"
deck card due set --board 1 --stack 2 --card 3 --value "2026-03-08T12:00:00Z"
deck card move --board 1 --from-stack 2 --to-stack 4 --card 3 --order 999
deck todo add --board 1 --stack 4 --card 3 --text "Follow up"
deck label create --board 1 --title "Blocked" --color FF7A66
deck comment create --card 3 --message "Need review"
deck attachment upload --board 1 --stack 4 --card 3 --file ./notes.txt
deck share list --board 1
deck config get
deck list create --board 1 --title "Backlog"
deck list reorder --board 1 --list 4 --order 10
```
