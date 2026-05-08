package deck

import "fmt"

type LookupError struct {
	Resource string
	Title    string
	BoardID  int64
	Matches  int
}

func (e LookupError) Error() string {
	scope := ""
	if e.BoardID != 0 {
		scope = fmt.Sprintf(" on board %d", e.BoardID)
	}
	if e.Matches == 0 {
		return fmt.Sprintf("%s title %q not found%s", e.Resource, e.Title, scope)
	}
	return fmt.Sprintf("%s title %q matched %d %ss%s; use id", e.Resource, e.Title, e.Matches, e.Resource, scope)
}
