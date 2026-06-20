// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

// ResponseContentType is the media type of an introspection response body
// (RFC 7662 §2.2).
const ResponseContentType = "application/json"

// maxRequestBytes bounds how much of a request body ParseRequest will read, so
// a hostile client cannot exhaust the responder's memory with an unbounded form
// body. A token plus a hint is tiny; 1 MiB is far above any legitimate request.
const maxRequestBytes = 1 << 20

// ParseRequest reads an RFC 7662 §2.1 introspection request from an incoming
// HTTP request on the authorization-server side. It is the producer-side
// counterpart to Client's request encoding.
//
// Unlike the liberal RequestFromValues codec, ParseRequest enforces the HTTP
// boundary rules: the body must be form-encoded
// (application/x-www-form-urlencoded) and the required token parameter must be
// present (§2.1). Either failure is a *ValidationError, naming the offending
// field so the caller can answer with HTTP 400. Caller authentication (§2.1) is
// the responder's concern and is not checked here.
//
// The request body is read through a 1 MiB limit to bound memory use on hostile
// input; an oversized body surfaces as a parse error.
func ParseRequest(r *http.Request) (*Request, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != FormContentType {
		return nil, &ValidationError{
			Field:   "Content-Type",
			Message: fmt.Sprintf("must be %s", FormContentType),
		}
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBytes)
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("introspection: parse form: %w", err)
	}
	// PostForm is body-only by design: a token must arrive in the request body,
	// never the URL query, where it would leak via logs, Referer, and history.
	// Do not switch this to r.Form.
	req := RequestFromValues(r.PostForm)
	if req.Token == "" {
		return nil, &ValidationError{Field: "token", Message: "required parameter is missing"}
	}
	return req, nil
}

// WriteResponse writes resp as the JSON body of an introspection response
// (RFC 7662 §2.2) with HTTP 200 and the application/json content type. It is the
// producer-side counterpart to Client decoding the response.
//
// The active member is always emitted, so the §2.2 "active REQUIRED" rule holds
// by construction. WriteResponse does not strip an inactive response of its
// other members: §2.3 advises a responder not to reveal them, but that is the
// responder's policy to apply to resp before calling, not something the library
// enforces.
func WriteResponse(w http.ResponseWriter, resp *Response) error {
	body, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("introspection: marshal response: %w", err)
	}
	w.Header().Set("Content-Type", ResponseContentType)
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("introspection: write response: %w", err)
	}
	return nil
}
