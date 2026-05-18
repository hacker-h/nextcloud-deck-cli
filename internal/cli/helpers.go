package cli

import "github.com/hacker-h/nextcloud-deck-api/internal/deck"

func boardSummary(board deck.Board) map[string]any {
	return map[string]any{"id": board.ID, "title": board.Title, "color": board.Color, "archived": board.Archived}
}

func stackSummary(stack deck.Stack) map[string]any {
	return map[string]any{"id": stack.ID, "boardId": stack.BoardID, "title": stack.Title, "order": stack.Order}
}

func cardSummary(card deck.Card) map[string]any {
	result := map[string]any{
		"id":          card.ID,
		"title":       card.Title,
		"stackId":     card.StackID,
		"order":       card.Order,
		"archived":    card.Archived,
		"description": card.Description,
	}
	if card.Duedate != nil {
		result["dueDate"] = *card.Duedate
	}
	if card.Startdate != nil {
		result["startDate"] = *card.Startdate
	}
	if card.Type != "" {
		result["type"] = card.Type
	}
	if card.Color != "" {
		result["color"] = card.Color
	}
	return result
}

func first(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func stringPtr(v string) *string { return &v }

func int64Ptr(v int64) *int64 { return &v }

func baseCardUpdate(card deck.Card, title string, description *string, due *string) deck.UpdateCardRequest {
	return deck.UpdateCardRequest{
		Title:       title,
		Description: description,
		Type:        first(card.Type, "plain"),
		Color:       card.Color,
		Order:       int64Ptr(card.Order),
		Duedate:     due,
		Startdate:   card.Startdate,
		Done:        card.Done,
		Owner:       card.Owner,
	}
}
