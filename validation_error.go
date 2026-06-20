// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"fmt"
)

// ErrValidation is the sentinel that every *ValidationError unwraps to, so a
// caller can match any validation failure with errors.Is(err, ErrValidation)
// or inspect the specifics with errors.As.
var ErrValidation = errors.New("introspection: validation failed")

// ValidationError reports a value that does not satisfy an RFC 7662 wire-shape
// requirement — a missing required parameter, an unsupported content type, and
// the like. It names the offending field so a caller can build a precise
// response.
type ValidationError struct {
	// Field is the parameter or member at fault, e.g. "token".
	Field string
	// Message explains the problem, lowercase and without trailing punctuation.
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("introspection: %s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error { return ErrValidation }
