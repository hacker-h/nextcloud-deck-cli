package deck

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) GetConfig(ctx context.Context) (map[string]any, error) {
	var data map[string]any
	err := c.doOCS(ctx, http.MethodGet, "/config", nil, &data)
	return data, err
}

func (c *Client) SetConfig(ctx context.Context, key string, value any) (any, error) {
	var data any
	err := c.doOCS(ctx, http.MethodPost, fmt.Sprintf("/config/%s", key), ConfigValueRequest{Value: value}, &data)
	return data, err
}
