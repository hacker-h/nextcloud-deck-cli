package deck

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) ListComments(ctx context.Context, cardID int64) ([]Comment, error) {
	var comments []Comment
	err := c.doOCS(ctx, http.MethodGet, fmt.Sprintf("/cards/%d/comments", cardID), nil, &comments)
	return comments, err
}

func (c *Client) CreateComment(ctx context.Context, cardID int64, message string) (Comment, error) {
	var comment Comment
	err := c.doOCS(ctx, http.MethodPost, fmt.Sprintf("/cards/%d/comments", cardID), CreateCommentRequest{Message: message}, &comment)
	return comment, err
}

func (c *Client) UpdateComment(ctx context.Context, cardID, commentID int64, message string) (Comment, error) {
	var comment Comment
	err := c.doOCS(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/comments/%d", cardID, commentID), UpdateCommentRequest{Message: message}, &comment)
	return comment, err
}

func (c *Client) DeleteComment(ctx context.Context, cardID, commentID int64) error {
	var ignored any
	return c.doOCS(ctx, http.MethodDelete, fmt.Sprintf("/cards/%d/comments/%d", cardID, commentID), nil, &ignored)
}
