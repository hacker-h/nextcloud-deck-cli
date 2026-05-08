package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestTodoMarkdownRoundTrip(t *testing.T) {
	desc := "Intro\n- [ ] first\n- [x] second"
	desc = addTodo(desc, "third")
	desc, err := setTodoState(desc, 1, true)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	todos := extractTodos(desc)
	if len(todos) != 3 {
		t.Fatalf("len(todos) = %d", len(todos))
	}
	if !todos[0].Checked || !todos[1].Checked || todos[2].Checked {
		t.Fatalf("unexpected todos: %#v", todos)
	}
}

func TestExtractTodos_EmptyDescription(t *testing.T) {
	if got := extractTodos(""); len(got) != 0 {
		t.Fatalf("extractTodos() = %#v", got)
	}
}

func TestExtractTodos_NoCheckboxes(t *testing.T) {
	if got := extractTodos("plain\n- item\n[x] nope"); len(got) != 0 {
		t.Fatalf("extractTodos() = %#v", got)
	}
}

func TestExtractTodos_MixedBullets(t *testing.T) {
	got := extractTodos("- [ ] dash\n* [x] star")
	want := []markdownTodo{{Index: 1, Text: "dash", Checked: false}, {Index: 2, Text: "star", Checked: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractTodos() = %#v", got)
	}
}

func TestExtractTodos_AllChecked(t *testing.T) {
	got := extractTodos("- [x] one\n\t* [x] two")
	if len(got) != 2 || !got[0].Checked || !got[1].Checked {
		t.Fatalf("extractTodos() = %#v", got)
	}
}

func TestExtractTodos_LeadingWhitespace(t *testing.T) {
	got := extractTodos("  - [ ] indented")
	if len(got) != 1 || got[0].Text != "indented" || got[0].Index != 1 {
		t.Fatalf("extractTodos() = %#v", got)
	}
}

func TestAddTodo_EmptyDescription(t *testing.T) {
	if got := addTodo(" \n", " task "); got != "- [ ] task" {
		t.Fatalf("addTodo() = %q", got)
	}
}

func TestAddTodo_TrailingNewline(t *testing.T) {
	if got := addTodo("Intro\n", "task"); got != "Intro\n- [ ] task" {
		t.Fatalf("addTodo() = %q", got)
	}
}

func TestAddTodo_NoTrailingNewline(t *testing.T) {
	if got := addTodo("Intro", "task"); got != "Intro\n- [ ] task" {
		t.Fatalf("addTodo() = %q", got)
	}
}

func TestSetTodoState_CheckFirst(t *testing.T) {
	got, err := setTodoState("- [ ] first\n- [ ] second", 1, true)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	if !strings.Contains(got, "- [x] first") || strings.Contains(got, "- [x] second") {
		t.Fatalf("setTodoState() = %q", got)
	}
}

func TestSetTodoState_UncheckLast(t *testing.T) {
	got, err := setTodoState("- [x] first\n- [x] second", 2, false)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	if got != "- [x] first\n- [ ] second" {
		t.Fatalf("setTodoState() = %q", got)
	}
}

func TestSetTodoState_OutOfRange(t *testing.T) {
	if _, err := setTodoState("- [ ] first", 2, true); err == nil || !strings.Contains(err.Error(), "todo index 2 not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestSetTodoState_PreserveIndentation(t *testing.T) {
	got, err := setTodoState("\t* [ ] nested", 1, true)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	if got != "\t- [x] nested" {
		t.Fatalf("setTodoState() = %q", got)
	}
}

func TestSetTodoState_PreserveNonTodoLines(t *testing.T) {
	input := "Intro\n- [ ] first\nOutro"
	got, err := setTodoState(input, 1, true)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	if got != "Intro\n- [x] first\nOutro" {
		t.Fatalf("setTodoState() = %q", got)
	}
}
