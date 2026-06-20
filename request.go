// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

// Request is an RFC 7662 §2.1 token introspection request: the token to
// introspect plus an optional hint about its type.
//
// The endpoint requires the caller to authenticate (§2.1); that is handled by
// the HTTP transport, not carried on this struct. See Client.
type Request struct {
	// Token is the string value of the token to introspect. REQUIRED (§2.1).
	Token string

	// TokenTypeHint is an optional hint about the type of Token, allowing the
	// server to optimise its lookup (§2.1). It is advisory: a server MAY ignore
	// it, and an incorrect hint MUST NOT change the result. Registered values
	// are TokenTypeHintAccessToken and TokenTypeHintRefreshToken, but any string
	// is accepted.
	TokenTypeHint string
}

// Token type hints registered for the introspection (and revocation) endpoints
// by RFC 7009 §2.1. TokenTypeHint accepts other values too; these are the
// well-known ones.
const (
	TokenTypeHintAccessToken  = "access_token"
	TokenTypeHintRefreshToken = "refresh_token"
)
