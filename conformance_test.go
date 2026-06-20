// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hstern/go-token-introspection/internal/specfixtures"
)

// TestConformanceResponseBothRoles drives every spec-derived response vector
// through both roles — the Client decodes it from a stub endpoint, the responder
// re-emits it with WriteResponse — and asserts the value survives and the bytes
// are stable.
func TestConformanceResponseBothRoles(t *testing.T) {
	for name, vec := range specfixtures.ResponseVectors {
		t.Run(name, func(t *testing.T) {
			c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, vec)
			})
			resp, err := c.Introspect(context.Background(), &Request{Token: "t"})
			if err != nil {
				t.Fatalf("Introspect: %v", err)
			}

			rec := httptest.NewRecorder()
			if err := WriteResponse(rec, resp); err != nil {
				t.Fatalf("WriteResponse: %v", err)
			}

			var roundTripped Response
			if err := json.Unmarshal(rec.Body.Bytes(), &roundTripped); err != nil {
				t.Fatalf("re-decode responder output: %v", err)
			}
			if !reflect.DeepEqual(*resp, roundTripped) {
				t.Errorf("both-roles round-trip drifted:\n client = %+v\n responder = %+v", resp, roundTripped)
			}
			again, err := json.Marshal(roundTripped)
			if err != nil {
				t.Fatal(err)
			}
			if rec.Body.String() != string(again) {
				t.Errorf("byte drift:\n responder = %s\n re-encoded = %s", rec.Body.String(), again)
			}
		})
	}
}

// TestConformanceInactiveNotError pins C1: a {"active":false} response is a
// normal answer, not an error.
func TestConformanceInactiveNotError(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, specfixtures.InactiveResponse)
	})
	resp, err := c.Introspect(context.Background(), &Request{Token: "t"})
	if err != nil {
		t.Fatalf("inactive token returned error: %v", err)
	}
	if resp.Active {
		t.Error("Active = true, want false")
	}
}

// TestConformanceRequestBothRoles drives each valid request form through the
// responder's ParseRequest and the client's EncodeForm and back, asserting the
// request survives the round-trip.
func TestConformanceRequestBothRoles(t *testing.T) {
	for name, form := range specfixtures.ValidRequests {
		t.Run(name, func(t *testing.T) {
			req, err := ParseRequest(formRequest(form, FormContentType))
			if err != nil {
				t.Fatalf("ParseRequest: %v", err)
			}
			req2, err := ParseRequest(formRequest(req.EncodeForm(), FormContentType))
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if !reflect.DeepEqual(req, req2) {
				t.Errorf("request round-trip drifted:\n in = %+v\n out = %+v", req, req2)
			}
		})
	}
}

// TestConformanceMissingTokenRejected pins C3: a request without token is
// rejected by the responder.
func TestConformanceMissingTokenRejected(t *testing.T) {
	_, err := ParseRequest(formRequest(specfixtures.MissingTokenRequest, FormContentType))
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "token" {
		t.Fatalf("err = %v, want *ValidationError on token", err)
	}
}
