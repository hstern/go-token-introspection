// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

// Package specfixtures holds RFC 7662 conformance vectors derived from the
// spec's example figures and prose MUSTs. RFC 7662 publishes no interop event
// and no machine-readable test vectors, so these stand in for an external
// suite: the introspection library is its own conformance source of truth.
//
// The vectors are consumed by the root package's conformance test, which drives
// both roles (client decode and responder parse/emit) through them. See the
// C1–C6 conformance list in .claude/spec-reference.md.
package specfixtures

// Figure4 is the active-token introspection response from RFC 7662 §2.2
// (Figure 4): the canonical "valid active" vector, carrying typed members of
// every kind plus a service-specific extension member.
const Figure4 = `{
  "active": true,
  "client_id": "l238j323ds-23ij4",
  "username": "jdoe",
  "scope": "read write dolphin",
  "sub": "Z5O3upPC88QrAjx00dis",
  "aud": "https://protected.example.net/resource",
  "iss": "https://server.example.com/",
  "exp": 1419356238,
  "iat": 1419350238,
  "extension_field": "twenty-seven"
}`

// InactiveResponse is the minimal inactive-token response (§2.2, §2.3): the only
// member is active=false. A conformant client treats it as a normal answer, not
// an error (C1).
const InactiveResponse = `{"active":false}`

// ResponseVectors are §2.2 response bodies a conformant codec must decode and
// re-encode without loss or byte drift. Keyed by a short description.
var ResponseVectors = map[string]string{
	"figure 4 active":   Figure4,
	"inactive":          InactiveResponse,
	"aud array":         `{"active":true,"aud":["a","b"]}`,                                    // C5
	"aud single string": `{"active":true,"aud":"a"}`,                                          // C5
	"numeric dates":     `{"active":true,"exp":1419356238,"iat":1419350238,"nbf":1419350238}`, // C6
	"extension member":  `{"active":true,"extension_field":"twenty-seven"}`,                   // C4
}

// ValidRequests are §2.1 form bodies a conformant parser must accept, keyed by
// description.
var ValidRequests = map[string]string{
	"token only":     "token=mF_9.B5f-4.1JqM",
	"token and hint": "token=mF_9.B5f-4.1JqM&token_type_hint=access_token",
}

// MissingTokenRequest is a §2.1 form body with no token; a conformant parser
// must reject it (C3).
const MissingTokenRequest = "token_type_hint=access_token"
