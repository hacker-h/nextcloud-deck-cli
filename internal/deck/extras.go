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
	var raw json.RawMessage
	if err := c.doOCS(ctx, http.MethodGet, "/overview/upcoming", nil, &raw); err != nil {
		return nil, err
	}
	return decodeUpcomingCards(raw)
}

func decodeUpcomingCards(raw json.RawMessage) ([]Card, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var cards []Card
	if err := json.Unmarshal(raw, &cards); err == nil {
		return cards, nil
	}

	var groups map[string][]Card
	if err := json.Unmarshal(raw, &groups); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		cards = append(cards, groups[key]...)
	}
	return cards, nil
}

func (c *Client) SearchCards(ctx context.Context, term string, limit int) ([]Card, error) {
	return c.SearchCardsCursor(ctx, term, limit, 0)
}

func (c *Client) SearchCardsCursor(ctx context.Context, term string, limit int, cursor int) ([]Card, error) {
	endpoint := fmt.Sprintf("/search?term=%s", url.QueryEscape(term))
	if limit > 0 {
		endpoint += fmt.Sprintf("&limit=%d", limit)
	}
	if cursor > 0 {
		endpoint += fmt.Sprintf("&cursor=%d", cursor)
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
	cardMap := map[int64]int64{}
	deferredCards := []map[string]any{}
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
			start := nullableString(cardData["startdate"])
			card, err := c.CreateCard(ctx, board.ID, stack.ID, CreateCardRequest{Title: toString(cardData["title"]), Type: firstString(toString(cardData["type"]), "plain"), Color: toString(cardData["color"]), Order: toInt64(cardData["order"]), Description: &description, Duedate: due, Startdate: start})
			if err != nil {
				continue
			}
			oldCardID := toInt64(cardData["id"])
			if oldCardID != 0 {
				cardMap[oldCardID] = card.ID
			}
			cardData["__newCardID"] = card.ID
			cardData["__newStackID"] = stack.ID
			deferredCards = append(deferredCards, cardData)
			for _, labelData := range asMapSlice(cardData["labels"]) {
				oldID := toInt64(labelData["id"])
				if newID, ok := labelMap[oldID]; ok {
					_ = c.AssignLabel(ctx, board.ID, stack.ID, card.ID, newID)
				}
			}
			for _, assignmentData := range asMapSlice(cardData["assignedUsers"]) {
				if userID := assignmentUserID(assignmentData); userID != "" {
					_, _ = c.AssignUser(ctx, board.ID, stack.ID, card.ID, userID)
				}
			}
			if cardData["done"] != nil {
				_, _ = c.MarkCardDone(ctx, card.ID)
			}
			for _, commentData := range asMapSlice(cardData["comments"]) {
				if message := toString(commentData["message"]); message != "" {
					_, _ = c.CreateComment(ctx, card.ID, message)
				}
			}
		}
	}
	for _, cardData := range deferredCards {
		cardID := toInt64(cardData["__newCardID"])
		stackID := toInt64(cardData["__newStackID"])
		for _, oldDependentID := range int64Slice(cardData["dependentCards"]) {
			if newDependentID, ok := cardMap[oldDependentID]; ok {
				_, _ = c.AssignDependentCard(ctx, board.ID, stackID, cardID, newDependentID)
			}
		}
	}
	for _, aclData := range asMapSlice(payload["acl"]) {
		if isTruthy(aclData["owner"]) {
			continue
		}
		participant := participantID(aclData["participant"])
		if participant == "" {
			continue
		}
		_, _ = c.CreateShare(ctx, board.ID, CreateACLRuleRequest{Type: int(toInt64(aclData["type"])), Participant: participant, PermissionEdit: isTruthy(aclData["permissionEdit"]), PermissionShare: isTruthy(aclData["permissionShare"]), PermissionManage: isTruthy(aclData["permissionManage"])})
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

func int64Slice(value any) []int64 {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	ids := make([]int64, 0, len(raw))
	for _, item := range raw {
		ids = append(ids, toInt64(item))
	}
	return ids
}

func assignmentUserID(data map[string]any) string {
	return participantID(data["participant"])
}

func participantID(value any) string {
	participant, ok := value.(map[string]any)
	if !ok {
		return toString(value)
	}
	for _, key := range []string{"uid", "primaryKey", "id"} {
		if value := toString(participant[key]); value != "" {
			return value
		}
	}
	return ""
}

func isTruthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case int:
		return v != 0
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}
