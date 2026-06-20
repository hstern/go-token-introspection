// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

// Package introspection implements RFC 7662, OAuth 2.0 Token Introspection: a
// typed request/response model and a small HTTP client and responder helpers
// for the introspection endpoint.
//
// A protected resource uses [Client] to ask an authorization server whether a
// token is active and to read its metadata:
//
//	c := introspection.NewClient("https://as.example.com/introspect",
//		introspection.WithBasicAuth("resource-server", "secret"))
//	resp, err := c.Introspect(ctx, &introspection.Request{Token: tok})
//	if err != nil {
//		// transport failure, non-200 status, or an undecodable body
//	}
//	if !resp.Active {
//		// the token is not active — a normal answer, not an error
//	}
//
// An inactive or unknown token comes back as a [Response] with Active false and
// a nil error (§2.3): it is a normal answer, not a failure. Errors are reserved
// for the request not completing.
//
// On the authorization-server side, [ParseRequest] and [WriteResponse] are
// à-la-carte helpers for an introspection endpoint bolted onto an existing
// server; the library ships no HTTP handler, leaving routing and caller
// authentication to the responder.
//
// The library depends only on the standard library. Client authentication
// beyond HTTP Basic ([WithBasicAuth]), along with TLS and timeouts, are
// transport concerns configured on the [net/http.Client] passed to
// [WithHTTPClient]. Verifying a signed (JWT) access token is out of scope — that
// is the local-validation alternative to introspection.
//
// Spec: https://www.rfc-editor.org/rfc/rfc7662.html
package introspection

// SpecVersion is the version of the specification this package targets.
const SpecVersion = "RFC 7662"
