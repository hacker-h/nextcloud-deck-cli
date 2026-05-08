package cli

import (
	"reflect"
	"testing"

	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

func TestBoardSummary(t *testing.T) {
	got := boardSummary(deck.Board{ID: 7, Title: "Board", Color: "ff0000", Archived: true})
	want := map[string]any{"id": int64(7), "title": "Board", "color": "ff0000", "archived": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("boardSummary() = %#v", got)
	}
}

func TestStackSummary(t *testing.T) {
	got := stackSummary(deck.Stack{ID: 2, BoardID: 7, Title: "Doing", Order: 3})
	want := map[string]any{"id": int64(2), "boardId": int64(7), "title": "Doing", "order": int64(3)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stackSummary() = %#v", got)
	}
}

func TestCardSummary(t *testing.T) {
	got := cardSummary(deck.Card{ID: 9, Title: "Card", StackID: 2, Order: 4, Archived: true, Description: "desc"})
	want := map[string]any{"id": int64(9), "title": "Card", "stackId": int64(2), "order": int64(4), "archived": true, "description": "desc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cardSummary() = %#v", got)
	}
}

func TestCardSummary_WithDueDate(t *testing.T) {
	due := "2026-05-09T10:00:00+00:00"
	got := cardSummary(deck.Card{ID: 9, Duedate: &due})
	if got["dueDate"] != due {
		t.Fatalf("dueDate = %#v", got["dueDate"])
	}
}

func TestCardSummary_NilDueDate(t *testing.T) {
	got := cardSummary(deck.Card{ID: 9})
	if _, ok := got["dueDate"]; ok {
		t.Fatalf("unexpected dueDate in %#v", got)
	}
}

func TestBaseCardUpdate(t *testing.T) {
	desc := "new desc"
	due := "2026-05-09"
	got := baseCardUpdate(deck.Card{Type: "markdown", Order: 12, Owner: map[string]any{"id": "u"}}, "New", &desc, &due)
	if got.Title != "New" || got.Description == nil || *got.Description != desc || got.Type != "markdown" || got.Order == nil || *got.Order != 12 || got.Duedate == nil || *got.Duedate != due {
		t.Fatalf("baseCardUpdate() = %#v", got)
	}
	if !reflect.DeepEqual(got.Owner, map[string]any{"id": "u"}) {
		t.Fatalf("Owner = %#v", got.Owner)
	}
}

func TestBaseCardUpdate_PreservesOwner(t *testing.T) {
	owner := []any{"owner"}
	got := baseCardUpdate(deck.Card{Order: 1, Owner: owner}, "Title", nil, nil)
	if got.Type != "plain" || got.Owner == nil || !reflect.DeepEqual(got.Owner, owner) {
		t.Fatalf("baseCardUpdate() = %#v", got)
	}
}

func TestFirst_WithValue(t *testing.T) {
	if got := first("value", "fallback"); got != "value" {
		t.Fatalf("first() = %q", got)
	}
}

func TestFirst_Fallback(t *testing.T) {
	if got := first("", "fallback"); got != "fallback" {
		t.Fatalf("first() = %q", got)
	}
}

func TestCoerceValue_True(t *testing.T) {
	if got := coerceValue(" true "); got != true {
		t.Fatalf("coerceValue() = %#v", got)
	}
}

func TestCoerceValue_False(t *testing.T) {
	if got := coerceValue("FALSE"); got != false {
		t.Fatalf("coerceValue() = %#v", got)
	}
}

func TestCoerceValue_Integer(t *testing.T) {
	if got := coerceValue("42"); got != int64(42) {
		t.Fatalf("coerceValue() = %#v (%T)", got, got)
	}
}

func TestCoerceValue_String(t *testing.T) {
	if got := coerceValue(" value "); got != "value" {
		t.Fatalf("coerceValue() = %#v", got)
	}
}
