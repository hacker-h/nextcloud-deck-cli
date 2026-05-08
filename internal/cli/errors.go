package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
	"github.com/hacker-h/nextcloud-deck-api/internal/deck"
)

type errorKind string

const (
	errorKindValidation errorKind = "validation"
	errorKindAuth       errorKind = "auth"
	errorKindNetwork    errorKind = "network"
	errorKindServer     errorKind = "server"
	errorKindAPI        errorKind = "api"
	errorKindInternal   errorKind = "internal"
)

type cliError struct {
	kind errorKind
	err  error
}

func (e cliError) Error() string {
	return e.err.Error()
}

func (e cliError) Unwrap() error {
	return e.err
}

func (e cliError) cliErrorKind() errorKind {
	return e.kind
}

func validationError(message string) error {
	return cliError{kind: errorKindValidation, err: errors.New(message)}
}

func validationf(format string, args ...any) error {
	return cliError{kind: errorKindValidation, err: fmt.Errorf(format, args...)}
}

func Main(args []string, stdout, stderr io.Writer) int {
	if err := Run(args, stdout, stderr); err != nil {
		writeError(stderr, err)
		return 1
	}
	return 0
}

func writeError(stderr io.Writer, err error) {
	_, _ = fmt.Fprintf(stderr, "error: %s: %s\n", classifyError(err), err)
}

func classifyError(err error) errorKind {
	var kinded interface{ cliErrorKind() errorKind }
	if errors.As(err, &kinded) {
		return kinded.cliErrorKind()
	}

	var missingEnv config.MissingEnvError
	if errors.As(err, &missingEnv) {
		return errorKindValidation
	}

	var apiErr deck.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden:
			return errorKindAuth
		case apiErr.StatusCode >= http.StatusInternalServerError:
			return errorKindServer
		default:
			return errorKindAPI
		}
	}

	var lookupErr deck.LookupError
	if errors.As(err, &lookupErr) {
		return errorKindValidation
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return errorKindNetwork
	}

	var netErr net.Error
	if errors.As(err, &netErr) || errors.Is(err, context.DeadlineExceeded) {
		return errorKindNetwork
	}

	if isFlagError(err) {
		return errorKindValidation
	}

	return errorKindInternal
}

func isFlagError(err error) bool {
	message := err.Error()
	return strings.HasPrefix(message, "flag provided but not defined:") ||
		strings.HasPrefix(message, "invalid value ") ||
		strings.HasPrefix(message, "parse error")
}
