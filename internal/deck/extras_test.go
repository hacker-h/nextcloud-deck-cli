package deck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func TestAsMapSlice_Nil(t *testing.T) {
	if got := asMapSlice(nil); got != nil {
		t.Fatalf("asMapSlice(nil) = %#v", got)
	}
}

func TestAsMapSlice_Empty(t *testing.T) {
	got := asMapSlice([]any{})
	if got == nil || len(got) != 0 {
		t.Fatalf("asMapSlice(empty) = %#v", got)
	}
}

func TestAsMapSlice_Valid(t *testing.T) {
	input := []any{map[string]any{"id": float64(1)}, map[string]any{"title": "x"}}
	got := asMapSlice(input)
	if len(got) != 2 || got[0]["id"] != float64(1) || got[1]["title"] != "x" {
		t.Fatalf("asMapSlice() = %#v", got)
	}
}

func TestAsMapSlice_MixedTypes(t *testing.T) {
	input := []any{map[string]any{"id": float64(1)}, "skip", nil, map[string]any{"id": float64(2)}}
	got := asMapSlice(input)
	if len(got) != 2 || got[0]["id"] != float64(1) || got[1]["id"] != float64(2) {
		t.Fatalf("asMapSlice() = %#v", got)
	}
}

func TestToInt64_Float64(t *testing.T) {
	if got := toInt64(float64(42.9)); got != 42 {
		t.Fatalf("toInt64() = %d", got)
	}
}

func TestToInt64_Int64(t *testing.T) {
	if got := toInt64(int64(42)); got != 42 {
		t.Fatalf("toInt64() = %d", got)
	}
}

func TestToInt64_Int(t *testing.T) {
	if got := toInt64(42); got != 42 {
		t.Fatalf("toInt64() = %d", got)
	}
}

func TestToInt64_JSONNumber(t *testing.T) {
	if got := toInt64(json.Number("42")); got != 42 {
		t.Fatalf("toInt64() = %d", got)
	}
}

func TestToInt64_String(t *testing.T) {
	if got := toInt64("42"); got != 0 {
		t.Fatalf("toInt64() = %d, want unsupported string to map to 0", got)
	}
}

func TestToInt64_Nil(t *testing.T) {
	if got := toInt64(nil); got != 0 {
		t.Fatalf("toInt64() = %d", got)
	}
}

func TestToString_String(t *testing.T) {
	if got := toString("value"); got != "value" {
		t.Fatalf("toString() = %q", got)
	}
}

func TestToString_NonString(t *testing.T) {
	if got := toString(123); got != "" {
		t.Fatalf("toString() = %q", got)
	}
}

func TestToString_Nil(t *testing.T) {
	if got := toString(nil); got != "" {
		t.Fatalf("toString() = %q", got)
	}
}

func TestNullableString_Nil(t *testing.T) {
	if got := nullableString(nil); got != nil {
		t.Fatalf("nullableString() = %#v", got)
	}
}

func TestNullableString_Empty(t *testing.T) {
	if got := nullableString(""); got != nil {
		t.Fatalf("nullableString() = %#v", got)
	}
}

func TestNullableString_NonEmpty(t *testing.T) {
	got := nullableString("value")
	if got == nil || *got != "value" {
		t.Fatalf("nullableString() = %#v", got)
	}
}

func TestFirstString_WithValue(t *testing.T) {
	if got := firstString("value", "fallback"); got != "value" {
		t.Fatalf("firstString() = %q", got)
	}
}

func TestFirstString_Fallback(t *testing.T) {
	if got := firstString("", "fallback"); got != "fallback" {
		t.Fatalf("firstString() = %q", got)
	}
}

func TestReadExportBoard_ValidJSON(t *testing.T) {
	path := t.TempDir() + "/board.json"
	writeTestFile(t, path, `{"title":"Board","stacks":[{"title":"Doing"}]}`)
	got, err := ReadExportBoard(path)
	if err != nil {
		t.Fatalf("ReadExportBoard() error = %v", err)
	}
	if got["title"] != "Board" || len(asMapSlice(got["stacks"])) != 1 {
		t.Fatalf("ReadExportBoard() = %#v", got)
	}
}

func TestReadExportBoard_InvalidJSON(t *testing.T) {
	path := t.TempDir() + "/board.json"
	writeTestFile(t, path, `{`)
	if _, err := ReadExportBoard(path); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestReadExportBoard_MissingFile(t *testing.T) {
	if _, err := ReadExportBoard(t.TempDir() + "/missing.json"); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestImportExportedBoard(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /index.php/apps/deck/api/v1.0/boards":
			_ = json.NewEncoder(w).Encode(Board{ID: 10, Title: "Imported", Color: "31cc7c", Labels: []Label{{ID: 100, Title: "Existing"}}})
		case "POST /index.php/apps/deck/api/v1.0/boards/10/labels":
			_ = json.NewEncoder(w).Encode(Label{ID: 101, Title: "New"})
		case "POST /index.php/apps/deck/api/v1.0/boards/10/stacks":
			_ = json.NewEncoder(w).Encode(Stack{ID: 20, BoardID: 10, Title: "Doing"})
		case "POST /index.php/apps/deck/api/v1.0/boards/10/stacks/20/cards":
			_ = json.NewEncoder(w).Encode(Card{ID: 30, Title: "Card", StackID: 20})
		case "PUT /index.php/apps/deck/api/v1.0/boards/10/stacks/20/cards/30/assignLabel":
			w.WriteHeader(http.StatusNoContent)
		case "GET /index.php/apps/deck/api/v1.0/boards/10":
			_ = json.NewEncoder(w).Encode(Board{ID: 10, Title: "Imported", Stacks: []Stack{{ID: 20, Cards: []Card{{ID: 30, Title: "Card"}}}}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(config.Config{BaseURL: server.URL, Username: "antonia", Password: "pw"})
	payload := map[string]any{
		"title": "Imported", "color": "31cc7c",
		"labels": []any{map[string]any{"id": float64(1), "title": "Existing", "color": "111111"}, map[string]any{"id": float64(2), "title": "New", "color": "222222"}},
		"stacks": []any{
			map[string]any{
				"id": float64(5), "title": "Doing", "order": float64(1),
				"cards": []any{
					map[string]any{
						"id": float64(6), "title": "Card", "type": "plain", "order": float64(1), "description": "desc",
						"labels": []any{map[string]any{"id": float64(1)}, map[string]any{"id": float64(2)}},
					},
				},
			},
		},
	}
	board, err := client.ImportExportedBoard(context.Background(), payload)
	if err != nil {
		t.Fatalf("ImportExportedBoard() error = %v", err)
	}
	if board.ID != 10 || board.Stacks[0].Cards[0].ID != 30 {
		t.Fatalf("board = %#v", board)
	}
	want := []string{
		"POST /index.php/apps/deck/api/v1.0/boards",
		"POST /index.php/apps/deck/api/v1.0/boards/10/labels",
		"POST /index.php/apps/deck/api/v1.0/boards/10/stacks",
		"POST /index.php/apps/deck/api/v1.0/boards/10/stacks/20/cards",
		"PUT /index.php/apps/deck/api/v1.0/boards/10/stacks/20/cards/30/assignLabel",
		"PUT /index.php/apps/deck/api/v1.0/boards/10/stacks/20/cards/30/assignLabel",
		"GET /index.php/apps/deck/api/v1.0/boards/10",
	}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %#v", requests)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := osWriteFile(path, contents); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func osWriteFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o600)
}
