package deck

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) FindStackByTitle(ctx context.Context, boardID int64, title string) (Stack, error) {
	stacks, err := c.GetStacks(ctx, boardID)
	if err != nil {
		return Stack{}, err
	}

	var match Stack
	matches := 0
	for _, stack := range stacks {
		if stack.Title == title {
			match = stack
			matches++
		}
	}
	if matches != 1 {
		return Stack{}, LookupError{Resource: "stack", Title: title, BoardID: boardID, Matches: matches}
	}
	return match, nil
}

func (c *Client) GetStacks(ctx context.Context, boardID int64) ([]Stack, error) {
	var stacks []Stack
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/stacks", boardID), nil, &stacks)
	return stacks, err
}

func (c *Client) GetArchivedStacks(ctx context.Context, boardID int64) ([]Stack, error) {
	var stacks []Stack
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/stacks/archived", boardID), nil, &stacks)
	return stacks, err
}

func (c *Client) DeleteStack(ctx context.Context, boardID, stackID int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/boards/%d/stacks/%d", boardID, stackID), nil, nil)
}
