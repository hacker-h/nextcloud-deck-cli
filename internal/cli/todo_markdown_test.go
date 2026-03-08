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
