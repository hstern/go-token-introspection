// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newServer starts a test introspection endpoint that runs handler and returns
// a Client pointed at it (plus extra options).
func newServer(t *testing.T, handler http.HandlerFunc, opts ...ClientOption) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, append([]ClientOption{WithHTTPClient(srv.Client())}, opts...)...)
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.WriteString(w, body); err != nil {
		t.Errorf("server write: %v", err)
	}
}

func TestIntrospectActive(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, figure4)
	})
	resp, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Active {
		t.Error("Active = false, want true")
	}
	if resp.ClientID != "l238j323ds-23ij4" {
		t.Errorf("ClientID = %q", resp.ClientID)
	}
}

// An inactive/unknown token is a normal 200 response, not an error (RFC 7662
// §2.3) — the single most-missed rule.
func TestIntrospectInactiveIsNotAnError(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"active":false}`)
	})
	resp, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if err != nil {
		t.Fatalf("inactive token returned error: %v", err)
	}
	if resp.Active {
		t.Error("Active = true, want false")
	}
}

func TestIntrospectSendsFormRequest(t *testing.T) {
	var gotMethod, gotCType, gotAccept, gotToken, gotHint string
	c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		gotToken = r.PostForm.Get("token")
		gotHint = r.PostForm.Get("token_type_hint")
		writeJSON(t, w, `{"active":true}`)
	})
	if _, err := c.Introspect(context.Background(), &Request{Token: "abc", TokenTypeHint: TokenTypeHintAccessToken}); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotCType != FormContentType {
		t.Errorf("Content-Type = %q, want %q", gotCType, FormContentType)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if gotToken != "abc" || gotHint != TokenTypeHintAccessToken {
		t.Errorf("form = token:%q hint:%q", gotToken, gotHint)
	}
}

func TestIntrospectBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	var gotOK bool
	c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotOK = r.BasicAuth()
		writeJSON(t, w, `{"active":true}`)
	}, WithBasicAuth("rs-client", "s3cr3t"))
	if _, err := c.Introspect(context.Background(), &Request{Token: "tok"}); err != nil {
		t.Fatal(err)
	}
	if !gotOK || gotUser != "rs-client" || gotPass != "s3cr3t" {
		t.Errorf("BasicAuth = (%q, %q, ok=%v)", gotUser, gotPass, gotOK)
	}
}

func TestIntrospectNoAuthByDefault(t *testing.T) {
	var hadAuth bool
	c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]
		writeJSON(t, w, `{"active":true}`)
	})
	if _, err := c.Introspect(context.Background(), &Request{Token: "tok"}); err != nil {
		t.Fatal(err)
	}
	if hadAuth {
		t.Error("Authorization header sent without WithBasicAuth")
	}
}

func TestIntrospectUnauthorized(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	_, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != http.StatusUnauthorized {
		t.Fatalf("err = %v, want *HTTPError with 401", err)
	}
}

func TestIntrospectUnexpectedStatus(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("err = %v, want ErrUnexpectedStatus", err)
	}
	if errors.Is(err, ErrUnauthorized) {
		t.Error("500 must not match ErrUnauthorized")
	}
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != http.StatusInternalServerError {
		t.Fatalf("err = %v, want *HTTPError with 500", err)
	}
}

func TestIntrospectMalformedBody(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{not json`)
	})
	_, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err = %v, want ErrInvalidResponse", err)
	}
}

func TestIntrospectTransportError(t *testing.T) {
	// Point at a server that is already closed: Do returns a transport error.
	srv := httptest.NewServer(http.NotFoundHandler())
	c := NewClient(srv.URL, WithHTTPClient(srv.Client()))
	srv.Close()
	_, err := c.Introspect(context.Background(), &Request{Token: "tok"})
	if err == nil {
		t.Fatal("want transport error, got nil")
	}
	// A transport failure is none of the response-level sentinels.
	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrUnexpectedStatus) || errors.Is(err, ErrInvalidResponse) {
		t.Errorf("transport error misclassified: %v", err)
	}
}

func TestIntrospectContextCanceled(t *testing.T) {
	c := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"active":true}`)
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Introspect(ctx, &Request{Token: "tok"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
