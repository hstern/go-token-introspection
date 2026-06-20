// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func formRequest(body, contentType string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/introspect", strings.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	return r
}

// A request the Client would build round-trips back to an equal Request through
// ParseRequest (Phase 3 encode <-> Phase 4 decode symmetry).
func TestParseRequestRoundTrip(t *testing.T) {
	in := &Request{Token: "mF_9.B5f-4.1JqM", TokenTypeHint: TokenTypeHintRefreshToken}
	out, err := ParseRequest(formRequest(in.EncodeForm(), FormContentType))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n in = %+v\nout = %+v", in, out)
	}
}

func TestParseRequestAcceptsCharsetParameter(t *testing.T) {
	if _, err := ParseRequest(formRequest("token=abc", FormContentType+"; charset=utf-8")); err != nil {
		t.Fatalf("charset content type rejected: %v", err)
	}
}

func TestParseRequestMissingToken(t *testing.T) {
	_, err := ParseRequest(formRequest("token_type_hint=access_token", FormContentType))
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "token" {
		t.Fatalf("err = %v, want *ValidationError on token", err)
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("err = %v, want to match ErrValidation", err)
	}
}

func TestParseRequestWrongContentType(t *testing.T) {
	_, err := ParseRequest(formRequest(`{"token":"abc"}`, "application/json"))
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "Content-Type" {
		t.Fatalf("err = %v, want *ValidationError on Content-Type", err)
	}
}

func TestParseRequestRejectsOversizedBody(t *testing.T) {
	huge := "token=" + strings.Repeat("a", (1<<20)+1)
	if _, err := ParseRequest(formRequest(huge, FormContentType)); err == nil {
		t.Fatal("oversized body accepted, want error")
	}
}

func TestWriteResponseRoundTrip(t *testing.T) {
	resp := &Response{Active: true, ClientID: "abc", Audience: Audience{"rs"}}
	rec := httptest.NewRecorder()
	if err := WriteResponse(rec, resp); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if got := res.Header.Get("Content-Type"); got != ResponseContentType {
		t.Errorf("Content-Type = %q, want %q", got, ResponseContentType)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
	var out Response
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*resp, out) {
		t.Errorf("decoded response drifted:\n in = %+v\nout = %+v", resp, out)
	}
}

func TestWriteResponseInactiveMinimal(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := WriteResponse(rec, &Response{Active: false}); err != nil {
		t.Fatal(err)
	}
	if got := rec.Body.String(); got != `{"active":false}` {
		t.Errorf("body = %s, want {\"active\":false}", got)
	}
}
