package deck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

const apiPrefix = "/index.php/apps/deck/api/v1.0"
const ocsPrefix = "/ocs/v2.php/apps/deck/api/v1.0"
const appPrefix = "/index.php/apps/deck"

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
}

func (e APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("deck api returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("deck api returned status %d: %s", e.StatusCode, e.Message)
}

func NewClient(cfg config.Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = config.DefaultTimeout
	}
	return &Client{
		baseURL:  cfg.BaseURL,
		username: cfg.Username,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				MaxConnsPerHost:       50,
				IdleConnTimeout:       config.DefaultTimeout,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (c *Client) GetBoards(ctx context.Context, details bool) ([]Board, error) {
	values := url.Values{}
	if details {
		values.Set("details", "true")
	}
	endpoint := "/boards"
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}
	var boards []Board
	err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &boards)
	return boards, err
}

func (c *Client) GetBoard(ctx context.Context, boardID int64) (Board, error) {
	var board Board
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d", boardID), nil, &board)
	return board, err
}

func (c *Client) GetStack(ctx context.Context, boardID, stackID int64) (Stack, error) {
	var stack Stack
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/stacks/%d", boardID, stackID), nil, &stack)
	return stack, err
}

func (c *Client) CreateStack(ctx context.Context, boardID int64, req CreateStackRequest) (Stack, error) {
	var stack Stack
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/stacks", boardID), req, &stack)
	return stack, err
}

func (c *Client) UpdateStack(ctx context.Context, boardID, stackID int64, req UpdateStackRequest) (Stack, error) {
	var stack Stack
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/stacks/%d", boardID, stackID), req, &stack)
	return stack, err
}

func (c *Client) CreateCard(ctx context.Context, boardID, stackID int64, req CreateCardRequest) (Card, error) {
	if req.Type == "" {
		req.Type = "plain"
	}
	var card Card
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/boards/%d/stacks/%d/cards", boardID, stackID), req, &card)
	return card, err
}

func (c *Client) GetCard(ctx context.Context, boardID, stackID, cardID int64) (Card, error) {
	var card Card
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/boards/%d/stacks/%d/cards/%d", boardID, stackID, cardID), nil, &card)
	return card, err
}

func (c *Client) UpdateCard(ctx context.Context, boardID, stackID, cardID int64, req UpdateCardRequest) (Card, error) {
	if req.Type == "" {
		req.Type = "plain"
	}
	var card Card
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/boards/%d/stacks/%d/cards/%d", boardID, stackID, cardID), req, &card)
	return card, err
}

func (c *Client) DeleteCard(ctx context.Context, boardID, stackID, cardID int64) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/boards/%d/stacks/%d/cards/%d", boardID, stackID, cardID), nil, nil)
}

func (c *Client) ReorderCard(ctx context.Context, boardID, stackID, cardID int64, req ReorderCardRequest) error {
	return c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/reorder", cardID), req, nil)
}

func (c *Client) ArchiveCard(ctx context.Context, boardID, stackID, cardID int64) (Card, error) {
	var card Card
	err := c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/archive", cardID), nil, &card)
	return card, err
}

func (c *Client) UnarchiveCard(ctx context.Context, boardID, stackID, cardID int64) (Card, error) {
	var card Card
	err := c.doAppJSON(ctx, http.MethodPut, fmt.Sprintf("/cards/%d/unarchive", cardID), nil, &card)
	return card, err
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, payload any, out any) error {
	return c.do(ctx, method, c.endpointURL(endpoint), payload, out, false, nil)
}

func (c *Client) doOCS(ctx context.Context, method, endpoint string, payload any, out any) error {
	return c.do(ctx, method, c.ocsURL(endpoint), payload, out, true, nil)
}

func (c *Client) doAppJSON(ctx context.Context, method, endpoint string, payload any, out any) error {
	return c.do(ctx, method, c.appURL(endpoint), payload, out, false, nil)
}

func (c *Client) do(ctx context.Context, method, endpoint string, payload any, out any, unwrapOCS bool, headers map[string]string) error {
	var body io.Reader
	if payload != nil {
		if raw, ok := payload.(*rawPayload); ok {
			body = raw.reader
		} else {
			data, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal request: %w", err)
			}
			body = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if unwrapOCS {
		if err := decodeOCSResponse(resp.Body, out); err != nil {
			return fmt.Errorf("decode ocs response: %w", err)
		}
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) endpointURL(endpoint string) string {
	u, _ := url.Parse(c.baseURL)
	endpointURL, _ := url.Parse(endpoint)
	u.Path = path.Join(u.Path, apiPrefix, strings.TrimPrefix(endpointURL.Path, "/"))
	u.RawQuery = endpointURL.RawQuery
	return u.String()
}

func (c *Client) ocsURL(endpoint string) string {
	u, _ := url.Parse(c.baseURL)
	endpointURL, _ := url.Parse(endpoint)
	u.Path = path.Join(u.Path, ocsPrefix, strings.TrimPrefix(endpointURL.Path, "/"))
	u.RawQuery = endpointURL.RawQuery
	return u.String()
}

func (c *Client) appURL(endpoint string) string {
	u, _ := url.Parse(c.baseURL)
	endpointURL, _ := url.Parse(endpoint)
	u.Path = path.Join(u.Path, appPrefix, strings.TrimPrefix(endpointURL.Path, "/"))
	u.RawQuery = endpointURL.RawQuery
	return u.String()
}

func (c *Client) doMultipart(ctx context.Context, method, endpoint string, fields map[string]string, fileField, filePath string, out any) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("write field %s: %w", key, err)
		}
	}
	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer file.Close()
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fileField, filepath.Base(filePath)))
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return fmt.Errorf("create multipart file: %w", err)
		}
		if _, err := io.Copy(part, file); err != nil {
			return fmt.Errorf("copy multipart file: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}
	url := endpoint
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = c.endpointURL(endpoint)
	}
	return c.do(ctx, method, url, &rawPayload{reader: &buffer}, out, false, map[string]string{"Content-Type": writer.FormDataContentType()})
}

func (c *Client) DownloadFile(ctx context.Context, endpoint, outPath string) error {
	return c.downloadToFile(ctx, c.endpointURL(endpoint), outPath)
}

func (c *Client) DownloadAppFile(ctx context.Context, endpoint, outPath string) error {
	return c.downloadToFile(ctx, c.appURL(endpoint), outPath)
}

func (c *Client) downloadToFile(ctx context.Context, rawURL, outPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/octet-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(resp)
	}
	file, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

func decodeAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return apiErrorFromBody(resp.StatusCode, body)
}

func apiErrorFromBody(statusCode int, body []byte) APIError {
	var apiErr APIError
	if len(body) > 0 && json.Unmarshal(body, &apiErr) == nil && (apiErr.StatusCode != 0 || apiErr.Message != "") {
		if apiErr.StatusCode == 0 {
			apiErr.StatusCode = statusCode
		}
		return apiErr
	}

	var ocsErr OCSResponse[json.RawMessage]
	if len(body) > 0 && json.Unmarshal(body, &ocsErr) == nil {
		if apiErr, ok := apiErrorFromOCSMeta(statusCode, ocsErr.OCS.Meta); ok {
			return apiErr
		}
	}

	return APIError{StatusCode: statusCode, Message: strings.TrimSpace(string(body))}
}

type rawPayload struct {
	reader io.Reader
}

func decodeOCSResponse(r io.Reader, out any) error {
	var wrapper OCSResponse[json.RawMessage]
	if err := json.NewDecoder(r).Decode(&wrapper); err != nil {
		return err
	}
	if apiErr, ok := apiErrorFromOCSMeta(http.StatusOK, wrapper.OCS.Meta); ok {
		return apiErr
	}
	if len(wrapper.OCS.Data) == 0 || string(wrapper.OCS.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(wrapper.OCS.Data, out); err != nil {
		return err
	}
	return nil
}

func apiErrorFromOCSMeta(fallbackStatusCode int, meta OCSMeta) (APIError, bool) {
	status := strings.ToLower(strings.TrimSpace(meta.Status))
	statusCode := meta.StatusCode
	if statusCode == 0 {
		statusCode = fallbackStatusCode
	}
	if status != "" && status != "ok" && status != "success" {
		return APIError{StatusCode: statusCode, Message: meta.Message}, true
	}
	if meta.StatusCode >= http.StatusBadRequest {
		return APIError{StatusCode: statusCode, Message: meta.Message}, true
	}
	return APIError{}, false
}
