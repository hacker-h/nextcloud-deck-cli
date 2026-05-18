package deck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) GetCapabilities(ctx context.Context) (map[string]any, error) {
	var data map[string]any
	err := c.do(ctx, http.MethodGet, c.nextcloudOCSURL("/cloud/capabilities?format=json"), nil, &data, true, nil)
	return data, err
}

func (c *Client) SearchSharees(ctx context.Context, term string) (map[string]any, error) {
	var data map[string]any
	endpoint := fmt.Sprintf("/apps/files_sharing/api/v1/sharees?format=json&lookup=false&perPage=20&itemType=0%%2C1%%2C7&search=%s", url.QueryEscape(term))
	err := c.do(ctx, http.MethodGet, c.nextcloudOCSURL(endpoint), nil, &data, true, nil)
	return data, err
}

func (c *Client) GetUser(ctx context.Context, userID string) (map[string]any, error) {
	var data map[string]any
	err := c.do(ctx, http.MethodGet, c.nextcloudOCSURL(fmt.Sprintf("/cloud/users/%s?format=json", url.PathEscape(userID))), nil, &data, true, nil)
	return data, err
}

func (c *Client) GetCardActivity(ctx context.Context, cardID int64) ([]Activity, error) {
	var data []Activity
	endpoint := fmt.Sprintf("/apps/activity/api/v2/activity/filter?format=json&object_type=deck_card&limit=50&since=-1&sort=asc&object_id=%d", cardID)
	err := c.do(ctx, http.MethodGet, c.nextcloudOCSURL(endpoint), nil, &data, true, nil)
	return data, err
}

type ActivityQuery struct {
	ObjectType string
	ObjectID   int64
	Limit      int
	Since      int64
	Sort       string
}

func (c *Client) GetActivity(ctx context.Context, query ActivityQuery) ([]Activity, error) {
	var data []Activity
	values := url.Values{"format": {"json"}}
	if query.ObjectType != "" {
		values.Set("object_type", query.ObjectType)
	}
	if query.ObjectID != 0 {
		values.Set("object_id", fmt.Sprint(query.ObjectID))
	}
	if query.Limit > 0 {
		values.Set("limit", fmt.Sprint(query.Limit))
	}
	if query.Since != 0 {
		values.Set("since", fmt.Sprint(query.Since))
	}
	if query.Sort != "" {
		values.Set("sort", query.Sort)
	}
	endpoint := "/apps/activity/api/v2/activity/filter?" + values.Encode()
	err := c.do(ctx, http.MethodGet, c.nextcloudOCSURL(endpoint), nil, &data, true, nil)
	return data, err
}

func (c *Client) nextcloudOCSURL(endpoint string) string {
	u, _ := url.Parse(c.baseURL)
	endpointURL, _ := url.Parse(endpoint)
	joinURLPath(u, "/ocs/v2.php", endpointURL)
	u.RawQuery = endpointURL.RawQuery
	return u.String()
}
