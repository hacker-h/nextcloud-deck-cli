package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
)

func (c *Client) CloneBoard(ctx context.Context, boardID int64, options map[string]bool) (Board, error) {
	var board Board
	err := c.doAppJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/clone", boardID), options, &board)
	return board, err
}

func (c *Client) ExportBoard(ctx context.Context, boardID int64, outPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.appURL(fmt.Sprintf("/boards/%d/export", boardID)), nil)
	if err != nil {
		return fmt.Errorf("build export request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform export request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(resp)
	}
	file, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

func (c *Client) ImportBoardFromFile(ctx context.Context, filePath string) (Board, error) {
	var board Board
	err := c.doMultipart(ctx, http.MethodPost, c.appURL("/boards/import"), nil, "file", filePath, &board)
	if err == nil {
		return board, nil
	}
	payload, readErr := ReadExportBoard(filePath)
	if readErr != nil {
		return Board{}, err
	}
	imported, importErr := c.ImportExportedBoard(ctx, payload)
	if importErr != nil {
		return Board{}, err
	}
	return imported, nil
}

func (c *Client) GetImportSystems(ctx context.Context) ([]string, error) {
	var systems []string
	err := c.doOCS(ctx, http.MethodGet, "/boards/import/getSystems", nil, &systems)
	return systems, err
}

func (c *Client) GetImportSchema(ctx context.Context, name string) (map[string]any, error) {
	var schema map[string]any
	err := c.doOCS(ctx, http.MethodGet, fmt.Sprintf("/boards/import/config/schema/%s", name), nil, &schema)
	return schema, err
}

func (c *Client) ImportBoard(ctx context.Context, req ImportRequest) (Board, error) {
	var board Board
	err := c.doOCS(ctx, http.MethodPost, "/boards/import", req, &board)
	return board, err
}

func (c *Client) CreateSession(ctx context.Context, boardID int64) (Session, error) {
	var session Session
	err := c.doOCS(ctx, http.MethodPut, "/session/create", map[string]any{"boardId": boardID}, &session)
	return session, err
}

func (c *Client) SyncSession(ctx context.Context, boardID int64, token string) error {
	var ignored []any
	return c.doOCS(ctx, http.MethodPost, "/session/sync", map[string]any{"boardId": boardID, "token": token}, &ignored)
}

func (c *Client) CloseSession(ctx context.Context, boardID int64, token string) error {
	var ignored []any
	return c.doOCS(ctx, http.MethodPost, "/session/close", map[string]any{"boardId": boardID, "token": token}, &ignored)
}

func (c *Client) UpcomingCards(ctx context.Context) ([]Card, error) {
	var cards []Card
	err := c.doOCS(ctx, http.MethodGet, "/overview/upcoming", nil, &cards)
	return cards, err
}

func (c *Client) SearchCards(ctx context.Context, term string, limit int) ([]Card, error) {
	endpoint := fmt.Sprintf("/search?term=%s", url.QueryEscape(term))
	if limit > 0 {
		endpoint += fmt.Sprintf("&limit=%d", limit)
	}
	var cards []Card
	err := c.doOCS(ctx, http.MethodGet, endpoint, nil, &cards)
	return cards, err
}

func (c *Client) CloneCard(ctx context.Context, cardID, targetStackID int64) (Card, error) {
	sourceCard, err := c.doReadCardByID(ctx, cardID)
	if err != nil {
		return Card{}, err
	}
	var card Card
	err = c.doOCS(ctx, http.MethodPost, fmt.Sprintf("/cards/%d/clone", cardID), map[string]any{"targetStackId": targetStackID}, &card)
	if err != nil {
		return Card{}, err
	}
	if card.ID != 0 {
		return card, nil
	}
	stacks, err := c.GetBoards(ctx, true)
	if err == nil {
		for _, board := range stacks {
			for _, stack := range board.Stacks {
				if stack.ID != targetStackID {
					continue
				}
				var newest Card
				for _, candidate := range stack.Cards {
					if candidate.Title == sourceCard.Title && candidate.ID != cardID && candidate.CreatedAt >= newest.CreatedAt {
						newest = candidate
					}
				}
				if newest.ID != 0 {
					return newest, nil
				}
			}
		}
	}
	return card, nil
}

func (c *Client) MarkCardDone(ctx context.Context, cardID int64) (Card, error) {
	var card Card
	err := c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/done", cardID), nil, &card)
	return card, err
}

func (c *Client) MarkCardUndone(ctx context.Context, cardID int64) (Card, error) {
	var card Card
	err := c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/undone", cardID), nil, &card)
	return card, err
}

func ReadExportBoard(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) ImportExportedBoard(ctx context.Context, payload map[string]any) (Board, error) {
	title, _ := payload["title"].(string)
	color, _ := payload["color"].(string)
	board, err := c.CreateBoard(ctx, BoardCreateRequest{Title: title, Color: firstString(color, "ff0000")})
	if err != nil {
		return Board{}, err
	}
	labelMap := map[int64]int64{}
	labels := asMapSlice(payload["labels"])
	for i, labelData := range labels {
		if i < len(board.Labels) {
			labelMap[toInt64(labelData["id"])] = board.Labels[i].ID
			continue
		}
		label, err := c.CreateLabel(ctx, board.ID, CreateLabelRequest{Title: toString(labelData["title"]), Color: toString(labelData["color"])})
		if err != nil {
			continue
		}
		labelMap[toInt64(labelData["id"])] = label.ID
	}
	stacks := asMapSlice(payload["stacks"])
	sort.Slice(stacks, func(i, j int) bool { return toInt64(stacks[i]["order"]) < toInt64(stacks[j]["order"]) })
	for _, stackData := range stacks {
		stack, err := c.CreateStack(ctx, board.ID, CreateStackRequest{Title: toString(stackData["title"]), Order: toInt64(stackData["order"])})
		if err != nil {
			continue
		}
		cards := asMapSlice(stackData["cards"])
		sort.Slice(cards, func(i, j int) bool { return toInt64(cards[i]["order"]) < toInt64(cards[j]["order"]) })
		for _, cardData := range cards {
			description := toString(cardData["description"])
			due := nullableString(cardData["duedate"])
			card, err := c.CreateCard(ctx, board.ID, stack.ID, CreateCardRequest{Title: toString(cardData["title"]), Type: firstString(toString(cardData["type"]), "plain"), Order: toInt64(cardData["order"]), Description: &description, Duedate: due})
			if err != nil {
				continue
			}
			for _, labelData := range asMapSlice(cardData["labels"]) {
				oldID := toInt64(labelData["id"])
				if newID, ok := labelMap[oldID]; ok {
					_ = c.AssignLabel(ctx, board.ID, stack.ID, card.ID, newID)
				}
			}
		}
	}
	return c.GetBoard(ctx, board.ID)
}

func asMapSlice(value any) []map[string]any {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if mapped, ok := item.(map[string]any); ok {
			items = append(items, mapped)
		}
	}
	return items
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	default:
		return 0
	}
}

func toString(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func nullableString(value any) *string {
	if value == nil {
		return nil
	}
	if s, ok := value.(string); ok && s != "" {
		return &s
	}
	return nil
}

func firstString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (c *Client) doReadCardByID(ctx context.Context, cardID int64) (Card, error) {
	var card Card
	err := c.doAppJSON(ctx, http.MethodGet, fmt.Sprintf("/cards/%d", cardID), nil, &card)
	return card, err
}
