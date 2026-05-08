package cli

import "testing"

func TestTodoMarkdownRoundTrip(t *testing.T) {
	description := addTodo("Intro", "first task")
	description = addTodo(description, "second task")
	todos := extractTodos(description)
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	if todos[0].Text != "first task" || todos[0].Checked {
		t.Fatalf("unexpected first todo: %#v", todos[0])
	}
	updated, err := setTodoState(description, 2, true)
	if err != nil {
		t.Fatalf("setTodoState() error = %v", err)
	}
	todos = extractTodos(updated)
	if !todos[1].Checked {
		t.Fatalf("expected second todo checked, got %#v", todos[1])
	}
}

// --- Todo markdown edge cases ---

func TestExtractTodos_EmptyDescription(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestExtractTodos_NoCheckboxes(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestExtractTodos_MixedBullets(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestExtractTodos_AllChecked(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestExtractTodos_LeadingWhitespace(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAddTodo_EmptyDescription(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAddTodo_TrailingNewline(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestAddTodo_NoTrailingNewline(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetTodoState_CheckFirst(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetTodoState_UncheckLast(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetTodoState_OutOfRange(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetTodoState_PreserveIndentation(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestSetTodoState_PreserveNonTodoLines(t *testing.T) {
	t.Skip("TODO: implement")
}
