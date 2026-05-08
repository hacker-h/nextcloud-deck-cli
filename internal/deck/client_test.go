package deck

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestCreateCardRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards/1/stacks/2/cards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("OCS-APIRequest"); got != "true" {
			t.Fatalf("OCS-APIRequest = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Basic "+base64.StdEncoding.EncodeToString([]byte("antonia:pw")) {
			t.Fatalf("Authorization = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["title"] != "Test card" || payload["type"] != "plain" {
			t.Fatalf("payload = %#v", payload)
		}
		writeJSON(w, Card{ID: 9, Title: "Test card", StackID: 2, Order: 999})
	}))
	defer server.Close()

	client := testClient(server.URL)
	card, err := client.CreateCard(context.Background(), 1, 2, CreateCardRequest{Title: "Test card", Order: 999})
	if err != nil {
		t.Fatalf("CreateCard() error = %v", err)
	}
	if card.ID != 9 {
		t.Fatalf("card.ID = %d", card.ID)
	}
}

func TestNewClientUsesConfiguredTimeout(t *testing.T) {
	client := NewClient(config.Config{BaseURL: "https://cloud.example.com", Username: "antonia", Password: "pw", Timeout: 2 * time.Minute})
	if client.httpClient.Timeout != 2*time.Minute {
		t.Fatalf("http client timeout = %s, want 2m", client.httpClient.Timeout)
	}
}

func TestNewClientDefaultsTimeout(t *testing.T) {
	client := NewClient(config.Config{BaseURL: "https://cloud.example.com", Username: "antonia", Password: "pw"})
	if client.httpClient.Timeout != config.DefaultTimeout {
		t.Fatalf("http client timeout = %s, want %s", client.httpClient.Timeout, config.DefaultTimeout)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"message":"title must be provided"}`))
	}))
	defer server.Close()

	client := testClient(server.URL)
	_, err := client.GetBoards(context.Background(), false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "title must be provided") {
		t.Fatalf("unexpected error: %v", err)
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
}

func TestAPIErrorDecodesOCSBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"ocs":{"meta":{"status":"failure","statuscode":500,"message":"maintenance"},"data":null}}`))
	}))
	defer server.Close()

	client := testClient(server.URL)
	_, err := client.GetBoards(context.Background(), false)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError || apiErr.Message != "maintenance" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}

func TestOCSMetaError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/config")
		writeOCS(w, OCSMeta{Status: "failure", StatusCode: 403, Message: "forbidden"}, nil)
	}))
	defer server.Close()

	client := testClient(server.URL)
	_, err := client.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Message != "forbidden" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}

// --- URL construction ---

func TestEndpointURL(t *testing.T) {
	client := testClient("https://cloud.example.com/nextcloud/")
	got := client.endpointURL("/boards/7?details=true")
	want := "https://cloud.example.com/nextcloud/index.php/apps/deck/api/v1.0/boards/7?details=true"
	if got != want {
		t.Fatalf("endpointURL() = %q", got)
	}
}

func TestOcsURL(t *testing.T) {
	client := testClient("https://cloud.example.com/root")
	got := client.ocsURL("/config?format=json")
	want := "https://cloud.example.com/root/ocs/v2.php/apps/deck/api/v1.0/config?format=json"
	if got != want {
		t.Fatalf("ocsURL() = %q", got)
	}
}

func TestAppURL(t *testing.T) {
	client := testClient("https://cloud.example.com/root/")
	got := client.appURL("/cards/9/archive")
	want := "https://cloud.example.com/root/index.php/apps/deck/cards/9/archive"
	if got != want {
		t.Fatalf("appURL() = %q", got)
	}
}

func TestNextcloudOCSURL(t *testing.T) {
	client := testClient("https://cloud.example.com/root")
	got := client.nextcloudOCSURL("/cloud/users/a%2Fb?format=json")
	want := "https://cloud.example.com/root/ocs/v2.php/cloud/users/a%2Fb?format=json"
	if got != want {
		t.Fatalf("nextcloudOCSURL() = %q", got)
	}
}

// --- HTTP plumbing ---

func TestDoJSON_GET_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/api/v1.0/ping")
		if got := r.URL.Query().Get("q"); got != "hello world" {
			t.Fatalf("query q = %q", got)
		}
		writeJSON(w, map[string]any{"ok": true})
	}))
	defer server.Close()
	var out map[string]any
	if err := testClient(server.URL).doJSON(context.Background(), http.MethodGet, "/ping?q=hello+world", nil, &out); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if out["ok"] != true {
		t.Fatalf("out = %#v", out)
	}
}

func TestDoJSON_POST_MarshalPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodPost, "/index.php/apps/deck/api/v1.0/ping")
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["title"] != "Card" {
			t.Fatalf("payload = %#v", payload)
		}
		writeJSON(w, payload)
	}))
	defer server.Close()
	var out map[string]any
	if err := testClient(server.URL).doJSON(context.Background(), http.MethodPost, "/ping", map[string]string{"title": "Card"}, &out); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
}

func TestDoJSON_NilOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodDelete, "/index.php/apps/deck/api/v1.0/ping")
		_, _ = w.Write([]byte(`{"ignored":true}`))
	}))
	defer server.Close()
	if err := testClient(server.URL).doJSON(context.Background(), http.MethodDelete, "/ping", nil, nil); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
}

func TestDoJSON_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := testClient("https://example.invalid").doJSON(ctx, http.MethodGet, "/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v", err)
	}
}

func TestDoJSON_Non2xx_NoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	defer server.Close()
	err := testClient(server.URL).doJSON(context.Background(), http.MethodGet, "/ping", nil, nil)
	var apiErr APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTeapot || apiErr.Message != "" {
		t.Fatalf("err = %#v", err)
	}
}

func TestDoJSON_Non2xx_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()
	err := testClient(server.URL).doJSON(context.Background(), http.MethodGet, "/ping", nil, nil)
	var apiErr APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadGateway || apiErr.Message != "not-json" {
		t.Fatalf("err = %#v", err)
	}
}

func TestDecodeOCSResponse_ValidData(t *testing.T) {
	var out map[string]any
	err := decodeOCSResponse(strings.NewReader(`{"ocs":{"meta":{"status":"ok","statuscode":200},"data":{"value":42}}}`), &out)
	if err != nil || out["value"].(float64) != 42 {
		t.Fatalf("out=%#v err=%v", out, err)
	}
}

func TestDecodeOCSResponse_NullData(t *testing.T) {
	out := map[string]any{"keep": true}
	err := decodeOCSResponse(strings.NewReader(`{"ocs":{"meta":{"status":"ok","statuscode":200},"data":null}}`), &out)
	if err != nil || out["keep"] != true {
		t.Fatalf("out=%#v err=%v", out, err)
	}
}

func TestDecodeOCSResponse_InvalidJSON(t *testing.T) {
	var out map[string]any
	if err := decodeOCSResponse(strings.NewReader(`{`), &out); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestDecodeAPIError_MissingStatusField(t *testing.T) {
	apiErr := apiErrorFromBody(http.StatusBadRequest, []byte(`{"message":"bad"}`))
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != "bad" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}

// --- Boards ---

func TestGetBoards(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards", []Board{{ID: 1, Title: "Board"}})
	defer server.Close()
	boards, err := testClient(server.URL).GetBoards(context.Background(), false)
	if err != nil || len(boards) != 1 || boards[0].ID != 1 {
		t.Fatalf("boards=%#v err=%v", boards, err)
	}
}

func TestGetBoardsWithDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards")
		if got := r.URL.Query().Get("details"); got != "true" {
			t.Fatalf("details = %q", got)
		}
		writeJSON(w, []Board{{ID: 1}})
	}))
	defer server.Close()
	if _, err := testClient(server.URL).GetBoards(context.Background(), true); err != nil {
		t.Fatalf("GetBoards(true) error = %v", err)
	}
}

func TestGetBoard(t *testing.T) {
	assertBoardResult(t, callJSONBoard(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7", func(c *Client) (Board, error) { return c.GetBoard(context.Background(), 7) }))
}
func TestCreateBoard(t *testing.T) {
	assertBoardResult(t, callJSONBoard(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards", func(c *Client) (Board, error) {
		return c.CreateBoard(context.Background(), BoardCreateRequest{Title: "Board", Color: "ff0000"})
	}))
}
func TestUpdateBoard(t *testing.T) {
	assertBoardResult(t, callJSONBoard(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7", func(c *Client) (Board, error) {
		return c.UpdateBoard(context.Background(), 7, BoardUpdateRequest{Title: "Board", Color: "ff0000"})
	}))
}
func TestDeleteBoard(t *testing.T) {
	assertNoContentCall(t, http.MethodDelete, "/index.php/apps/deck/api/v1.0/boards/7", func(c *Client) error { return c.DeleteBoard(context.Background(), 7) })
}
func TestRestoreBoard(t *testing.T) {
	assertBoardResult(t, callJSONBoard(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards/7/undo_delete", func(c *Client) (Board, error) { return c.RestoreBoard(context.Background(), 7) }))
}

// --- Stacks ---

func TestGetStacks(t *testing.T) {
	assertStacksResult(t, callJSONStacks(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/stacks", func(c *Client) ([]Stack, error) { return c.GetStacks(context.Background(), 7) }))
}
func TestGetArchivedStacks(t *testing.T) {
	assertStacksResult(t, callJSONStacks(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/stacks/archived", func(c *Client) ([]Stack, error) { return c.GetArchivedStacks(context.Background(), 7) }))
}
func TestGetStack(t *testing.T) {
	assertStackResult(t, callJSONStack(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2", func(c *Client) (Stack, error) { return c.GetStack(context.Background(), 7, 2) }))
}
func TestCreateStack(t *testing.T) {
	assertStackResult(t, callJSONStack(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards/7/stacks", func(c *Client) (Stack, error) {
		return c.CreateStack(context.Background(), 7, CreateStackRequest{Title: "Doing"})
	}))
}
func TestUpdateStack(t *testing.T) {
	assertStackResult(t, callJSONStack(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2", func(c *Client) (Stack, error) {
		return c.UpdateStack(context.Background(), 7, 2, UpdateStackRequest{Title: "Doing"})
	}))
}
func TestDeleteStack(t *testing.T) {
	assertNoContentCall(t, http.MethodDelete, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2", func(c *Client) error { return c.DeleteStack(context.Background(), 7, 2) })
}

// --- Cards ---

func TestGetCard(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", func(c *Client) (Card, error) { return c.GetCard(context.Background(), 7, 2, 9) }))
}
func TestUpdateCard(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", func(c *Client) (Card, error) {
		return c.UpdateCard(context.Background(), 7, 2, 9, UpdateCardRequest{Title: "Card"})
	}))
}
func TestDeleteCard(t *testing.T) {
	assertNoContentCall(t, http.MethodDelete, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9", func(c *Client) error { return c.DeleteCard(context.Background(), 7, 2, 9) })
}
func TestReorderCard(t *testing.T) {
	assertNoContentCall(t, http.MethodPut, "/index.php/apps/deck/cards/9/reorder", func(c *Client) error {
		return c.ReorderCard(context.Background(), 7, 2, 9, ReorderCardRequest{Order: 1, StackID: 2})
	})
}
func TestArchiveCard(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodPut, "/index.php/apps/deck/cards/9/archive", func(c *Client) (Card, error) { return c.ArchiveCard(context.Background(), 7, 2, 9) }))
}
func TestUnarchiveCard(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodPut, "/index.php/apps/deck/cards/9/unarchive", func(c *Client) (Card, error) { return c.UnarchiveCard(context.Background(), 7, 2, 9) }))
}

func TestListCards(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2", Stack{ID: 2, Cards: []Card{{ID: 9}}})
	defer server.Close()
	cards, err := testClient(server.URL).ListCards(context.Background(), 7, 2)
	if err != nil || len(cards) != 1 || cards[0].ID != 9 {
		t.Fatalf("cards=%#v err=%v", cards, err)
	}
}

func TestAssignLabel(t *testing.T) {
	assertNoContentCall(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/assignLabel", func(c *Client) error { return c.AssignLabel(context.Background(), 7, 2, 9, 4) })
}
func TestRemoveLabel(t *testing.T) {
	assertNoContentCall(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/stacks/2/cards/9/removeLabel", func(c *Client) error { return c.RemoveLabel(context.Background(), 7, 2, 9, 4) })
}
func TestAssignUser(t *testing.T) {
	server := jsonRouteServer(t, http.MethodPost, "/index.php/apps/deck/cards/9/assign", Assignment{ID: 5})
	defer server.Close()
	assignment, err := testClient(server.URL).AssignUser(context.Background(), 7, 2, 9, "alice")
	if err != nil || assignment.ID != 5 {
		t.Fatalf("assignment=%#v err=%v", assignment, err)
	}
}
func TestUnassignUser(t *testing.T) {
	assertNoContentCall(t, http.MethodPut, "/index.php/apps/deck/cards/9/unassign", func(c *Client) error { return c.UnassignUser(context.Background(), 7, 2, 9, "alice") })
}

// --- Labels ---

func TestGetLabel(t *testing.T) {
	assertLabelResult(t, callJSONLabel(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7/labels/4", func(c *Client) (Label, error) { return c.GetLabel(context.Background(), 7, 4) }))
}
func TestListLabels(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7", Board{ID: 7, Labels: []Label{{ID: 4}}})
	defer server.Close()
	labels, err := testClient(server.URL).ListLabels(context.Background(), 7)
	if err != nil || len(labels) != 1 || labels[0].ID != 4 {
		t.Fatalf("labels=%#v err=%v", labels, err)
	}
}
func TestCreateLabel(t *testing.T) {
	assertLabelResult(t, callJSONLabel(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards/7/labels", func(c *Client) (Label, error) {
		return c.CreateLabel(context.Background(), 7, CreateLabelRequest{Title: "Bug"})
	}))
}
func TestUpdateLabel(t *testing.T) {
	assertLabelResult(t, callJSONLabel(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/labels/4", func(c *Client) (Label, error) {
		return c.UpdateLabel(context.Background(), 7, 4, UpdateLabelRequest{Title: "Bug"})
	}))
}
func TestDeleteLabel(t *testing.T) {
	assertNoContentCall(t, http.MethodDelete, "/index.php/apps/deck/api/v1.0/boards/7/labels/4", func(c *Client) error { return c.DeleteLabel(context.Background(), 7, 4) })
}

// --- Comments ---

func TestListComments(t *testing.T) {
	comments, err := callOCSComments(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/cards/9/comments", func(c *Client) ([]Comment, error) { return c.ListComments(context.Background(), 9) })
	if err != nil || len(comments) != 1 {
		t.Fatalf("comments=%#v err=%v", comments, err)
	}
}
func TestCreateComment(t *testing.T) {
	comment, err := callOCSComment(t, http.MethodPost, "/ocs/v2.php/apps/deck/api/v1.0/cards/9/comments", func(c *Client) (Comment, error) { return c.CreateComment(context.Background(), 9, "hello") })
	if err != nil || comment.ID != 6 {
		t.Fatalf("comment=%#v err=%v", comment, err)
	}
}
func TestUpdateComment(t *testing.T) {
	comment, err := callOCSComment(t, http.MethodPut, "/ocs/v2.php/apps/deck/api/v1.0/cards/9/comments/6", func(c *Client) (Comment, error) { return c.UpdateComment(context.Background(), 9, 6, "hello") })
	if err != nil || comment.ID != 6 {
		t.Fatalf("comment=%#v err=%v", comment, err)
	}
}
func TestDeleteComment(t *testing.T) {
	assertOCSNoContentCall(t, http.MethodDelete, "/ocs/v2.php/apps/deck/api/v1.0/cards/9/comments/6", func(c *Client) error { return c.DeleteComment(context.Background(), 9, 6) })
}

// --- Attachments ---

func TestListAttachments(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/cards/9/attachments", []Attachment{{ID: 8}})
	defer server.Close()
	attachments, err := testClient(server.URL).ListAttachments(context.Background(), 7, 2, 9)
	if err != nil || len(attachments) != 1 {
		t.Fatalf("attachments=%#v err=%v", attachments, err)
	}
}

func TestUploadAttachment(t *testing.T) {
	file := t.TempDir() + "/note.txt"
	writeClientFile(t, file, "hello")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodPost, "/index.php/apps/deck/cards/9/attachment")
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.FormValue("type"); got != "file" {
			t.Fatalf("type = %q", got)
		}
		if _, header, err := r.FormFile("file"); err != nil || header.Filename != "note.txt" {
			t.Fatalf("file header=%#v err=%v", header, err)
		}
		writeJSON(w, Attachment{ID: 8})
	}))
	defer server.Close()
	attachment, err := testClient(server.URL).UploadAttachment(context.Background(), 7, 2, 9, file)
	if err != nil || attachment.ID != 8 {
		t.Fatalf("attachment=%#v err=%v", attachment, err)
	}
}

func TestDeleteAttachment(t *testing.T) {
	assertAttachmentRefSequence(t, http.MethodDelete, "/index.php/apps/deck/cards/9/attachment/deck_file:8", func(c *Client) error { return c.DeleteAttachment(context.Background(), 7, 2, 9, 8) })
}
func TestRestoreAttachment(t *testing.T) {
	attachment, err := callAttachmentRefSequence(t, http.MethodGet, "/index.php/apps/deck/cards/9/attachment/deck_file:8/restore", func(c *Client) (Attachment, error) { return c.RestoreAttachment(context.Background(), 7, 2, 9, 8) })
	if err != nil || attachment.ID != 8 {
		t.Fatalf("attachment=%#v err=%v", attachment, err)
	}
}
func TestDownloadAttachment(t *testing.T) {
	out := t.TempDir() + "/out.txt"
	assertAttachmentDownload(t, out)
	data, err := os.ReadFile(out)
	if err != nil || string(data) != "contents" {
		t.Fatalf("data=%q err=%v", string(data), err)
	}
}
func TestAttachmentRef(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/cards/9/attachments", []Attachment{{ID: 8}})
	defer server.Close()
	ref, err := testClient(server.URL).attachmentRef(context.Background(), 7, 2, 9, 8)
	if err != nil || ref != "deck_file:8" {
		t.Fatalf("ref=%q err=%v", ref, err)
	}
}
func TestAttachmentRef_NotFound(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/cards/9/attachments", []Attachment{{ID: 7}})
	defer server.Close()
	_, err := testClient(server.URL).attachmentRef(context.Background(), 7, 2, 9, 8)
	if err == nil || !strings.Contains(err.Error(), "attachment 8 not found") {
		t.Fatalf("err=%v", err)
	}
}

// --- Shares ---

func TestListShares(t *testing.T) {
	server := jsonRouteServer(t, http.MethodGet, "/index.php/apps/deck/api/v1.0/boards/7", Board{ID: 7, ACL: []ACLRule{{ID: 3}}})
	defer server.Close()
	shares, err := testClient(server.URL).ListShares(context.Background(), 7)
	if err != nil || len(shares) != 1 || shares[0].ID != 3 {
		t.Fatalf("shares=%#v err=%v", shares, err)
	}
}
func TestCreateShare_ArrayResponse(t *testing.T) {
	server := jsonRouteServer(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards/7/acl", []ACLRule{{ID: 3}})
	defer server.Close()
	shares, err := testClient(server.URL).CreateShare(context.Background(), 7, CreateACLRuleRequest{Participant: "alice"})
	if err != nil || len(shares) != 1 {
		t.Fatalf("shares=%#v err=%v", shares, err)
	}
}
func TestCreateShare_SingleResponse(t *testing.T) {
	server := jsonRouteServer(t, http.MethodPost, "/index.php/apps/deck/api/v1.0/boards/7/acl", ACLRule{ID: 3})
	defer server.Close()
	shares, err := testClient(server.URL).CreateShare(context.Background(), 7, CreateACLRuleRequest{Participant: "alice"})
	if err != nil || len(shares) != 1 || shares[0].ID != 3 {
		t.Fatalf("shares=%#v err=%v", shares, err)
	}
}
func TestUpdateShare(t *testing.T) {
	assertNoContentCall(t, http.MethodPut, "/index.php/apps/deck/api/v1.0/boards/7/acl/3", func(c *Client) error {
		return c.UpdateShare(context.Background(), 7, 3, UpdateACLRuleRequest{PermissionEdit: true})
	})
}
func TestDeleteShare(t *testing.T) {
	assertNoContentCall(t, http.MethodDelete, "/index.php/apps/deck/api/v1.0/boards/7/acl/3", func(c *Client) error { return c.DeleteShare(context.Background(), 7, 3) })
}

// --- Extras ---

func TestCloneBoard(t *testing.T) {
	assertBoardResult(t, callJSONBoard(t, http.MethodPost, "/index.php/apps/deck/boards/7/clone", func(c *Client) (Board, error) {
		return c.CloneBoard(context.Background(), 7, map[string]bool{"withCards": true})
	}))
}
func TestCloneCard(t *testing.T) {
	card, err := callCloneCardSequence(t)
	if err != nil || card.ID != 10 {
		t.Fatalf("card=%#v err=%v", card, err)
	}
}
func TestMarkCardDone(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodPut, "/index.php/apps/deck/cards/9/done", func(c *Client) (Card, error) { return c.MarkCardDone(context.Background(), 9) }))
}
func TestMarkCardUndone(t *testing.T) {
	assertCardResult(t, callJSONCard(t, http.MethodPut, "/index.php/apps/deck/cards/9/undone", func(c *Client) (Card, error) { return c.MarkCardUndone(context.Background(), 9) }))
}
func TestSearchCards(t *testing.T) {
	cards, err := callOCSCards(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/search", "term=hello+world&limit=5", func(c *Client) ([]Card, error) { return c.SearchCards(context.Background(), "hello world", 5) })
	if err != nil || len(cards) != 1 {
		t.Fatalf("cards=%#v err=%v", cards, err)
	}
}
func TestUpcomingCards(t *testing.T) {
	cards, err := callOCSCards(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/overview/upcoming", "", func(c *Client) ([]Card, error) { return c.UpcomingCards(context.Background()) })
	if err != nil || len(cards) != 1 {
		t.Fatalf("cards=%#v err=%v", cards, err)
	}
}
func TestExportBoard(t *testing.T) {
	out := t.TempDir() + "/board.json"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/boards/7/export")
		_, _ = w.Write([]byte(`{"title":"Board"}`))
	}))
	defer server.Close()
	if err := testClient(server.URL).ExportBoard(context.Background(), 7, out); err != nil {
		t.Fatalf("ExportBoard() error = %v", err)
	}
	data, _ := os.ReadFile(out)
	if string(data) != `{"title":"Board"}` {
		t.Fatalf("data = %q", data)
	}
}
func TestImportBoardFromFile(t *testing.T) {
	file := t.TempDir() + "/board.json"
	writeClientFile(t, file, `{"title":"Board"}`)
	server := jsonRouteServer(t, http.MethodPost, "/index.php/apps/deck/boards/import", Board{ID: 7})
	defer server.Close()
	board, err := testClient(server.URL).ImportBoardFromFile(context.Background(), file)
	if err != nil || board.ID != 7 {
		t.Fatalf("board=%#v err=%v", board, err)
	}
}
func TestGetImportSystems(t *testing.T) {
	systems, err := callOCSStrings(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/boards/import/getSystems", func(c *Client) ([]string, error) { return c.GetImportSystems(context.Background()) })
	if err != nil || len(systems) != 1 || systems[0] != "deck" {
		t.Fatalf("systems=%#v err=%v", systems, err)
	}
}
func TestGetImportSchema(t *testing.T) {
	data, err := callOCSMap(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/boards/import/config/schema/deck", func(c *Client) (map[string]any, error) { return c.GetImportSchema(context.Background(), "deck") })
	if err != nil || data["type"] != "object" {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}

// --- Sessions ---

func TestCreateSession(t *testing.T) {
	session, err := callOCSSession(t, http.MethodPut, "/ocs/v2.php/apps/deck/api/v1.0/session/create", func(c *Client) (Session, error) { return c.CreateSession(context.Background(), 7) })
	if err != nil || session.Token != "token" {
		t.Fatalf("session=%#v err=%v", session, err)
	}
}
func TestSyncSession(t *testing.T) {
	assertOCSNoContentCall(t, http.MethodPost, "/ocs/v2.php/apps/deck/api/v1.0/session/sync", func(c *Client) error { return c.SyncSession(context.Background(), 7, "token") })
}
func TestCloseSession(t *testing.T) {
	assertOCSNoContentCall(t, http.MethodPost, "/ocs/v2.php/apps/deck/api/v1.0/session/close", func(c *Client) error { return c.CloseSession(context.Background(), 7, "token") })
}

// --- Config API ---

func TestGetConfig(t *testing.T) {
	data, err := callOCSMap(t, http.MethodGet, "/ocs/v2.php/apps/deck/api/v1.0/config", func(c *Client) (map[string]any, error) { return c.GetConfig(context.Background()) })
	if err != nil || data["calendar"] != true {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}
func TestSetConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodPost, "/ocs/v2.php/apps/deck/api/v1.0/config/calendar")
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload["value"] != true {
			t.Fatalf("payload=%#v", payload)
		}
		writeOCS(w, OCSMeta{Status: "ok", StatusCode: 200}, map[string]any{"ok": true})
	}))
	defer server.Close()
	data, err := testClient(server.URL).SetConfig(context.Background(), "calendar", true)
	if err != nil || data.(map[string]any)["ok"] != true {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}

// --- Nextcloud ---

func TestGetCapabilities(t *testing.T) {
	data, err := callOCSMap(t, http.MethodGet, "/ocs/v2.php/cloud/capabilities", func(c *Client) (map[string]any, error) { return c.GetCapabilities(context.Background()) })
	if err != nil || data["calendar"] != true {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}
func TestSearchSharees(t *testing.T) {
	data, err := callOCSMapQuery(t, http.MethodGet, "/ocs/v2.php/apps/files_sharing/api/v1/sharees", "format=json&lookup=false&perPage=20&itemType=0%2C1%2C7&search=alice", func(c *Client) (map[string]any, error) { return c.SearchSharees(context.Background(), "alice") })
	if err != nil || data["calendar"] != true {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}
func TestGetUser(t *testing.T) {
	data, err := callOCSMap(t, http.MethodGet, "/ocs/v2.php/cloud/users/alice", func(c *Client) (map[string]any, error) { return c.GetUser(context.Background(), "alice") })
	if err != nil || data["calendar"] != true {
		t.Fatalf("data=%#v err=%v", data, err)
	}
}
func TestGetCardActivity(t *testing.T) {
	activities, err := callOCSActivities(t, http.MethodGet, "/ocs/v2.php/apps/activity/api/v2/activity/filter", "format=json&object_type=deck_card&limit=50&since=-1&sort=asc&object_id=9", func(c *Client) ([]Activity, error) { return c.GetCardActivity(context.Background(), 9) })
	if err != nil || len(activities) != 1 {
		t.Fatalf("activities=%#v err=%v", activities, err)
	}
}

func testClient(baseURL string) *Client {
	return NewClient(config.Config{BaseURL: baseURL, Username: "antonia", Password: "pw"})
}

func assertRequest(t *testing.T, r *http.Request, method, requestPath string) {
	t.Helper()
	if r.Method != method || r.URL.Path != requestPath {
		t.Fatalf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, requestPath)
	}
	if got := r.Header.Get("OCS-APIRequest"); got != "true" {
		t.Fatalf("OCS-APIRequest = %q", got)
	}
	if got := r.Header.Get("Accept"); got == "" {
		t.Fatal("missing Accept header")
	}
	if got := r.Header.Get("Authorization"); got != "Basic "+base64.StdEncoding.EncodeToString([]byte("antonia:pw")) {
		t.Fatalf("Authorization = %q", got)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
func writeOCS(w http.ResponseWriter, meta OCSMeta, data any) {
	writeJSON(w, map[string]any{"ocs": map[string]any{"meta": meta, "data": data}})
}

func jsonRouteServer(t *testing.T, method, requestPath string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, method, requestPath)
		writeJSON(w, response)
	}))
}

func assertNoContentCall(t *testing.T, method, requestPath string, call func(*Client) error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, method, requestPath)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	if err := call(testClient(server.URL)); err != nil {
		t.Fatalf("call error = %v", err)
	}
}
func assertOCSNoContentCall(t *testing.T, method, requestPath string, call func(*Client) error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, method, requestPath)
		writeOCS(w, OCSMeta{Status: "ok", StatusCode: 200}, []any{})
	}))
	defer server.Close()
	if err := call(testClient(server.URL)); err != nil {
		t.Fatalf("call error = %v", err)
	}
}

func callJSONBoard(t *testing.T, method, path string, call func(*Client) (Board, error)) Board {
	t.Helper()
	server := jsonRouteServer(t, method, path, Board{ID: 7, Title: "Board"})
	defer server.Close()
	board, err := call(testClient(server.URL))
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
	return board
}
func callJSONStack(t *testing.T, method, path string, call func(*Client) (Stack, error)) Stack {
	t.Helper()
	server := jsonRouteServer(t, method, path, Stack{ID: 2, BoardID: 7, Title: "Doing"})
	defer server.Close()
	stack, err := call(testClient(server.URL))
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
	return stack
}
func callJSONStacks(t *testing.T, method, path string, call func(*Client) ([]Stack, error)) []Stack {
	t.Helper()
	server := jsonRouteServer(t, method, path, []Stack{{ID: 2, BoardID: 7}})
	defer server.Close()
	stacks, err := call(testClient(server.URL))
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
	return stacks
}
func callJSONCard(t *testing.T, method, path string, call func(*Client) (Card, error)) Card {
	t.Helper()
	server := jsonRouteServer(t, method, path, Card{ID: 9, Title: "Card"})
	defer server.Close()
	card, err := call(testClient(server.URL))
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
	return card
}
func callJSONLabel(t *testing.T, method, path string, call func(*Client) (Label, error)) Label {
	t.Helper()
	server := jsonRouteServer(t, method, path, Label{ID: 4, Title: "Bug"})
	defer server.Close()
	label, err := call(testClient(server.URL))
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
	return label
}
func assertBoardResult(t *testing.T, board Board) {
	t.Helper()
	if board.ID != 7 {
		t.Fatalf("board=%#v", board)
	}
}
func assertStackResult(t *testing.T, stack Stack) {
	t.Helper()
	if stack.ID != 2 {
		t.Fatalf("stack=%#v", stack)
	}
}
func assertStacksResult(t *testing.T, stacks []Stack) {
	t.Helper()
	if len(stacks) != 1 || stacks[0].ID != 2 {
		t.Fatalf("stacks=%#v", stacks)
	}
}
func assertCardResult(t *testing.T, card Card) {
	t.Helper()
	if card.ID != 9 {
		t.Fatalf("card=%#v", card)
	}
}
func assertLabelResult(t *testing.T, label Label) {
	t.Helper()
	if label.ID != 4 {
		t.Fatalf("label=%#v", label)
	}
}

func ocsRouteServer(t *testing.T, method, path, query string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, method, path)
		if query != "" && r.URL.RawQuery != query {
			t.Fatalf("query = %q, want %q", r.URL.RawQuery, query)
		}
		writeOCS(w, OCSMeta{Status: "ok", StatusCode: 200}, response)
	}))
}
func callOCSComment(t *testing.T, method, path string, call func(*Client) (Comment, error)) (Comment, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, "", Comment{ID: 6, Message: "hello"})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSComments(t *testing.T, method, path string, call func(*Client) ([]Comment, error)) ([]Comment, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, "", []Comment{{ID: 6}})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSCards(t *testing.T, method, path, query string, call func(*Client) ([]Card, error)) ([]Card, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, query, []Card{{ID: 9}})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSStrings(t *testing.T, method, path string, call func(*Client) ([]string, error)) ([]string, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, "", []string{"deck"})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSMap(t *testing.T, method, path string, call func(*Client) (map[string]any, error)) (map[string]any, error) {
	t.Helper()
	return callOCSMapQuery(t, method, path, "", call)
}
func callOCSMapQuery(t *testing.T, method, path, query string, call func(*Client) (map[string]any, error)) (map[string]any, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, query, map[string]any{"calendar": true, "type": "object"})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSSession(t *testing.T, method, path string, call func(*Client) (Session, error)) (Session, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, "", Session{Token: "token"})
	defer server.Close()
	return call(testClient(server.URL))
}
func callOCSActivities(t *testing.T, method, path, query string, call func(*Client) ([]Activity, error)) ([]Activity, error) {
	t.Helper()
	server := ocsRouteServer(t, method, path, query, []Activity{{ActivityID: 1}})
	defer server.Close()
	return call(testClient(server.URL))
}

func assertAttachmentRefSequence(t *testing.T, finalMethod, finalPath string, call func(*Client) error) {
	t.Helper()
	_, err := callAttachmentRefSequence(t, finalMethod, finalPath, func(c *Client) (Attachment, error) { return Attachment{}, call(c) })
	if err != nil {
		t.Fatalf("call error = %v", err)
	}
}
func callAttachmentRefSequence(t *testing.T, finalMethod, finalPath string, call func(*Client) (Attachment, error)) (Attachment, error) {
	t.Helper()
	seen := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen++
		switch seen {
		case 1:
			assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/cards/9/attachments")
			writeJSON(w, []Attachment{{ID: 8}})
		case 2:
			assertRequest(t, r, finalMethod, finalPath)
			writeJSON(w, Attachment{ID: 8})
		default:
			t.Fatalf("unexpected extra request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	result, err := call(testClient(server.URL))
	if seen != 2 {
		t.Fatalf("seen requests = %d", seen)
	}
	return result, err
}
func assertAttachmentDownload(t *testing.T, out string) {
	t.Helper()
	seen := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen++
		switch seen {
		case 1:
			assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/cards/9/attachments")
			writeJSON(w, []Attachment{{ID: 8}})
		case 2:
			assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/cards/9/attachment/deck_file:8")
			_, _ = io.WriteString(w, "contents")
		default:
			t.Fatalf("unexpected request")
		}
	}))
	defer server.Close()
	if err := testClient(server.URL).DownloadAttachment(context.Background(), 7, 2, 9, 8, out); err != nil {
		t.Fatalf("DownloadAttachment() error = %v", err)
	}
}

func callCloneCardSequence(t *testing.T) (Card, error) {
	t.Helper()
	seen := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen++
		switch seen {
		case 1:
			assertRequest(t, r, http.MethodGet, "/index.php/apps/deck/cards/9")
			writeJSON(w, Card{ID: 9, Title: "Source"})
		case 2:
			assertRequest(t, r, http.MethodPost, "/ocs/v2.php/apps/deck/api/v1.0/cards/9/clone")
			writeOCS(w, OCSMeta{Status: "ok", StatusCode: 200}, Card{ID: 10, Title: "Source"})
		default:
			t.Fatalf("unexpected request")
		}
	}))
	defer server.Close()
	card, err := testClient(server.URL).CloneCard(context.Background(), 9, 2)
	if seen != 2 {
		t.Fatalf("seen requests = %d", seen)
	}
	return card, err
}

func writeClientFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
