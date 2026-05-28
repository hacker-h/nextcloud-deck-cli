---
name: nextcloud-deck-cli
description: Use when working with the nextcloud-deck-cli repository or helping users operate the deck CLI for Nextcloud Deck auth, board/list/card workflows, profile selection, safe verification, and avoiding unintended live writes.
---

# nextcloud-deck-cli

Use the repo's hand-rolled Go CLI patterns. Keep changes small, additive, gofmt'd, and covered by `go test ./...`.

## User Flows

- Set up auth with `deck auth setup`; use `deck auth setup --profile NAME` for named profiles.
- Select profiles with `deck --profile NAME ...` or `DECK_PROFILE=NAME deck ...`.
- List boards with `deck board list`; find exact board titles with `deck board find --title TEXT`.
- List stacks with friendly board selectors: `deck list --board <id-or-title>`, `deck list board <id-or-title>`, `deck stack --board <id-or-title>`, or `deck stacks --board <id-or-title>`.
- Use numeric IDs when titles are ambiguous. Board title selectors resolve exact title, case-insensitive exact title, then unique case-insensitive substring.
- Find a stack/list with `deck list find --board <id-or-title> --title <list-title>`.
- Create a card in Daily after resolving IDs: `deck card create --board <board-id> --stack <stack-id> --title TEXT --due 2026-05-28T17:00:00Z`.

## Verification

- Prefer read-only commands for smoke checks: `deck board list`, `deck list --board <id-or-title>`, `deck list find --board <id-or-title> --title TEXT`.
- Run repository checks with `gofmt` on changed Go files and `go test ./...`.
- Do not run live write commands such as create, update, delete, move, archive, import, or auth setup unless the user explicitly asks.
- Never print passwords, app passwords, environment dumps, or saved config contents.
