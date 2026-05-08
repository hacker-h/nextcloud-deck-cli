package deck

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) FindLabelByTitle(ctx context.Context, boardID int64, title string) (Label, error) {
	labels, err := c.ListLabels(ctx, boardID)
	if err != nil {
		return Label{}, err
	}

	var match Label
	matches := 0
	for _, label := range labels {
		if label.Title == title {
			match = label
			matches++
		}
	}
	if matches != 1 {
		return Label{}, LookupError{Resource: "label", Title: title, BoardID: boardID, Matches: matches}
	}
	return match, nil
}

func (c *Client) GetLabel(ctx context.Context, boardID, labelID int64) (Label, error) {
	var label Label
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/labels/%d", boardID, labelID), nil, &label)
	return label, err
}

func (c *Client) ListLabels(ctx context.Context, boardID int64) ([]Label, error) {
	board, err := c.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return board.Labels, nil
}

func (c *Client) CreateLabel(ctx context.Context, boardID int64, req CreateLabelRequest) (Label, error) {
	var label Label
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/labels", boardID), req, &label)
	return label, err
}

func (c *Client) UpdateLabel(ctx context.Context, boardID, labelID int64, req UpdateLabelRequest) (Label, error) {
	var label Label
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/labels/%d", boardID, labelID), req, &label)
	return label, err
}

func (c *Client) DeleteLabel(ctx context.Context, boardID, labelID int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/boards/%d/labels/%d", boardID, labelID), nil, nil)
}
