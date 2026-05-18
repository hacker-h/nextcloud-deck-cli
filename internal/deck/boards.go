package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) FindBoardByTitle(ctx context.Context, title string) (Board, error) {
	boards, err := c.GetBoards(ctx, false)
	if err != nil {
		return Board{}, err
	}

	var match Board
	matches := 0
	for _, board := range boards {
		if board.Title == title {
			match = board
			matches++
		}
	}
	if matches != 1 {
		return Board{}, LookupError{Resource: "board", Title: title, Matches: matches}
	}
	return match, nil
}

func (c *Client) CreateBoard(ctx context.Context, req BoardCreateRequest) (Board, error) {
	var board Board
	err := c.doJSON(ctx, http.MethodPost, "/boards", req, &board)
	return board, err
}

func (c *Client) UpdateBoard(ctx context.Context, boardID int64, req BoardUpdateRequest) (Board, error) {
	var board Board
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d", boardID), req, &board)
	return board, err
}

func (c *Client) DeleteBoard(ctx context.Context, boardID int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/boards/%d", boardID), nil, nil)
}

func (c *Client) RestoreBoard(ctx context.Context, boardID int64) (Board, error) {
	var board Board
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/undo_delete", boardID), nil, &board)
	return board, err
}

func (c *Client) ListShares(ctx context.Context, boardID int64) ([]ACLRule, error) {
	board, err := c.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return board.ACL, nil
}

func (c *Client) CreateShare(ctx context.Context, boardID int64, req CreateACLRuleRequest) ([]ACLRule, error) {
	var raw json.RawMessage
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/acl", boardID), req, &raw)
	if err != nil {
		return nil, err
	}
	var rules []ACLRule
	if err := json.Unmarshal(raw, &rules); err == nil {
		return rules, nil
	}
	var rule ACLRule
	if err := json.Unmarshal(raw, &rule); err == nil {
		return []ACLRule{rule}, nil
	}
	return nil, fmt.Errorf("decode share response: %s", string(raw))
}

func (c *Client) UpdateShare(ctx context.Context, boardID, aclID int64, req UpdateACLRuleRequest) error {
	return c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/acl/%d", boardID, aclID), req, nil)
}

func (c *Client) DeleteShare(ctx context.Context, boardID, aclID int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/boards/%d/acl/%d", boardID, aclID), nil, nil)
}

func (c *Client) GetBoardPermissions(ctx context.Context, boardID int64) (map[string]bool, error) {
	permissions := map[string]bool{}
	err := c.doAppJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/permissions", boardID), nil, &permissions)
	return permissions, err
}

func (c *Client) LeaveBoard(ctx context.Context, boardID int64) error {
	return c.doAppJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/leave", boardID), nil, nil)
}

func (c *Client) TransferBoardOwner(ctx context.Context, boardID int64, newOwner string) (map[string]any, error) {
	result := map[string]any{}
	err := c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/transferOwner", boardID), map[string]string{"newOwner": newOwner}, &result)
	return result, err
}
