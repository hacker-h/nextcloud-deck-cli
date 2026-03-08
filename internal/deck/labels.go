package deck

import (
	"context"
	"fmt"
	"net/http"
)

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
