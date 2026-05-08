package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunBoardCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.php/apps/deck/api/v1.0/boards" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"title":"Test Board","color":"ff0000","archived":false}`))
	}))
	defer server.Close()

	t.Setenv("NEXTCLOUD_BASE_URL", server.URL)
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	var stdout, stderr bytes.Buffer
	if err := Run([]string{"board", "create", "--title", "Test Board"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v; stderr=%s", err, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout=%s", err, stdout.String())
	}
	if payload["id"].(float64) != 7 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"help"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run(help) error = %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected help output")
	}
}

// --- Run basics ---

func TestRun_NoArgs(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_UnknownCommand(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Board subcommands ---

func TestRun_BoardList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardGet_MissingFlag(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardArchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardUnarchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardClone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardExport(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImport(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardRestore(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImportSystems(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_BoardImportSchema(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- List subcommands ---

func TestRun_ListList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListArchived(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListRename(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListReorder(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ListDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Card subcommands ---

func TestRun_CardList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardClone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardMove(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardReorder(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardArchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUnarchive(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUndone(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardRename(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDescribe(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueSet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardDueClear(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardAssignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardUnassignUser(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardAssignLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CardRemoveLabel(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Label subcommands ---

func TestRun_LabelList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_LabelDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Comment subcommands ---

func TestRun_CommentList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_CommentDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Attachment subcommands ---

func TestRun_AttachmentList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentUpload(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentDownload(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_AttachmentRestore(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Share subcommands ---

func TestRun_ShareList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareUpdate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ShareDelete(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Config subcommands ---

func TestRun_ConfigGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ConfigSet(t *testing.T) {
	t.Skip("TODO: implement")
}

// --- Other subcommands ---

func TestRun_SearchCards(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_OverviewUpcoming(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_Capabilities(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_UserSearch(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_UserGet(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_ActivityCard(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionCreate(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionSync(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_SessionClose(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoList(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoAdd(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoCheck(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestRun_TodoUncheck(t *testing.T) {
	t.Skip("TODO: implement")
}
