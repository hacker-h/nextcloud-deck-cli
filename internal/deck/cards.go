package deck

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListCards(ctx context.Context, boardID, stackID int64) ([]Card, error) {
	stack, err := c.GetStack(ctx, boardID, stackID)
	if err != nil {
		return nil, err
	}
	return stack.Cards, nil
}

func (c *Client) AssignLabel(ctx context.Context, boardID, stackID, cardID, labelID int64) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/stacks/%d/cards/%d/assignLabel", boardID, stackID, cardID), AssignLabelRequest{LabelID: labelID}, nil)
}

func (c *Client) RemoveLabel(ctx context.Context, boardID, stackID, cardID, labelID int64) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/stacks/%d/cards/%d/removeLabel", boardID, stackID, cardID), AssignLabelRequest{LabelID: labelID}, nil)
}

func (c *Client) AssignUser(ctx context.Context, boardID, stackID, cardID int64, userID string) (Assignment, error) {
	var assignment Assignment
	err := c.doAppJSON(ctx, http.MethodPost, fmt.Sprintf("/cards/%d/assign", cardID), map[string]any{"userId": userID, "type": 0}, &assignment)
	return assignment, err
}

func (c *Client) UnassignUser(ctx context.Context, boardID, stackID, cardID int64, userID string) error {
	return c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/unassign", cardID), map[string]any{"userId": userID, "type": 0}, nil)
}
