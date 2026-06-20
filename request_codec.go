// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import "net/url"

// FormContentType is the media type of an introspection request body (§2.1).
const FormContentType = "application/x-www-form-urlencoded"

// FormValues encodes the request as url.Values. token is always set; an empty
// TokenTypeHint is omitted.
func (r *Request) FormValues() url.Values {
	v := url.Values{}
	v.Set("token", r.Token)
	if r.TokenTypeHint != "" {
		v.Set("token_type_hint", r.TokenTypeHint)
	}
	return v
}

// EncodeForm returns the application/x-www-form-urlencoded request body.
func (r *Request) EncodeForm() string {
	return r.FormValues().Encode()
}

// RequestFromValues builds a Request from parsed form values. It is liberal
// (Postel's law): it copies whatever is present without validating that token
// is non-empty — enforcement of the §2.1 "token REQUIRED" rule happens at the
// HTTP boundary in ParseRequest. The companion to FormValues.
func RequestFromValues(v url.Values) *Request {
	return &Request{
		Token:         v.Get("token"),
		TokenTypeHint: v.Get("token_type_hint"),
	}
}
