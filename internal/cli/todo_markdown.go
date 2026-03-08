package cli

import (
	"fmt"
	"strings"
)

type markdownTodo struct {
	Index   int    `json:"index"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

func extractTodos(description string) []markdownTodo {
	lines := strings.Split(description, "\n")
	items := make([]markdownTodo, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		checked := strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ")
		open := strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ")
		if !checked && !open {
			continue
		}
		text := strings.TrimSpace(trimmed[6:])
		items = append(items, markdownTodo{Index: len(items) + 1, Text: text, Checked: checked})
	}
	return items
}

func addTodo(description, text string) string {
	line := "- [ ] " + strings.TrimSpace(text)
	if strings.TrimSpace(description) == "" {
		return line
	}
	if strings.HasSuffix(description, "\n") {
		return description + line
	}
	return description + "\n" + line
}

func setTodoState(description string, index int, checked bool) (string, error) {
	lines := strings.Split(description, "\n")
	current := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") || strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ") {
			current++
			if current == index {
				prefix := "- [ ] "
				if checked {
					prefix = "- [x] "
				}
				text := strings.TrimSpace(trimmed[6:])
				leading := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				lines[i] = leading + prefix + text
				return strings.Join(lines, "\n"), nil
			}
		}
	}
	return "", fmt.Errorf("todo index %d not found", index)
}
