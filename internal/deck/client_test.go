package deck

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		if payload["title"] != "Test card" {
			t.Fatalf("title = %#v", payload["title"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":9,"title":"Test card","stackId":2,"order":999,"archived":false}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	card, err := client.CreateCard(context.Background(), 1, 2, CreateCardRequest{Title: "Test card", Order: 999})
	if err != nil {
		t.Fatalf("CreateCard() error = %v", err)
	}
	if card.ID != 9 {
		t.Fatalf("card.ID = %d", card.ID)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"message":"title must be provided"}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
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

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
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
		if r.URL.Path != "/ocs/v2.php/apps/deck/api/v1.0/config" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ocs":{"meta":{"status":"failure","statuscode":403,"message":"forbidden"},"data":null}}`))
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
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
	t.Skip("TODO: implement")
}

func TestOcsURL(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAppURL(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNextcloudOCSURL(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- HTTP plumbing ---

func TestDoJSON_GET_Success(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDoJSON_POST_MarshalPayload(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDoJSON_NilOutput(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDoJSON_ContextCancelled(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDoJSON_Non2xx_NoBody(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDoJSON_Non2xx_InvalidJSON(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDecodeOCSResponse_ValidData(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDecodeOCSResponse_NullData(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDecodeOCSResponse_InvalidJSON(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDecodeAPIError_MissingStatusField(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Boards ---

func TestGetBoards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetBoardsWithDetails(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRestoreBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Stacks ---

func TestGetStacks(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetArchivedStacks(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetStack(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateStack(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateStack(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteStack(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Cards ---

func TestGetCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestReorderCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestArchiveCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUnarchiveCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestListCards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAssignLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRemoveLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAssignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUnassignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Labels ---

func TestGetLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestListLabels(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Comments ---

func TestListComments(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateComment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateComment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteComment(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Attachments ---

func TestListAttachments(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUploadAttachment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteAttachment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRestoreAttachment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDownloadAttachment(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAttachmentRef(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAttachmentRef_NotFound(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Shares ---

func TestListShares(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateShare_ArrayResponse(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCreateShare_SingleResponse(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpdateShare(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestDeleteShare(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Extras ---

func TestCloneBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCloneCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestMarkCardDone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestMarkCardUndone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSearchCards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestUpcomingCards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestExportBoard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestImportBoardFromFile(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetImportSystems(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetImportSchema(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Sessions ---

func TestCreateSession(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSyncSession(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestCloseSession(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Config API ---

func TestGetConfig(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetConfig(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Nextcloud ---

func TestGetCapabilities(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSearchSharees(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestGetCardActivity(t *testing.T) {
	t.Skip("TODO: implement")
}
