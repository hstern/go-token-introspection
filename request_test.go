// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"net/url"
	"reflect"
	"testing"
)

func TestRequestEncodeForm(t *testing.T) {
	got := (&Request{Token: "mF_9.B5f-4.1JqM", TokenTypeHint: TokenTypeHintAccessToken}).EncodeForm()
	want := "token=mF_9.B5f-4.1JqM&token_type_hint=access_token"
	if got != want {
		t.Fatalf("EncodeForm() = %q, want %q", got, want)
	}
}

func TestRequestEncodeFormOmitsEmptyHint(t *testing.T) {
	got := (&Request{Token: "abc"}).EncodeForm()
	want := "token=abc"
	if got != want {
		t.Fatalf("EncodeForm() = %q, want %q", got, want)
	}
}

func TestRequestFromValuesRoundTrip(t *testing.T) {
	in := &Request{Token: "abc def/+", TokenTypeHint: TokenTypeHintRefreshToken}
	out := RequestFromValues(in.FormValues())
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n in = %+v\nout = %+v", in, out)
	}
}

// RequestFromValues is liberal: a missing token yields an empty Token rather
// than an error (the §2.1 "token REQUIRED" rule is enforced at the HTTP
// boundary by ParseRequest, not here).
func TestRequestFromValuesLiberalOnMissingToken(t *testing.T) {
	out := RequestFromValues(url.Values{})
	if out.Token != "" || out.TokenTypeHint != "" {
		t.Fatalf("RequestFromValues(empty) = %+v, want zero Request", out)
	}
}
