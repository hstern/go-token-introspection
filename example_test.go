// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	introspection "github.com/hstern/go-token-introspection"
)

// Introspect a token against an authorization server. An inactive token is a
// normal answer (Active false, nil error), not an error.
func ExampleClient_Introspect() {
	// Stand in for a real authorization server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"active":true,"client_id":"l238j323ds-23ij4","scope":"read write"}`)
	}))
	defer srv.Close()

	c := introspection.NewClient(srv.URL,
		introspection.WithHTTPClient(srv.Client()),
		introspection.WithBasicAuth("resource-server", "secret"),
	)
	resp, err := c.Introspect(context.Background(), &introspection.Request{Token: "mF_9.B5f-4.1JqM"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Active, resp.ClientID, resp.Scopes())
	// Output: true l238j323ds-23ij4 [read write]
}

// Parse an introspection request on the authorization-server side.
func ExampleParseRequest() {
	r := httptest.NewRequest(http.MethodPost, "/introspect",
		// In a real handler this body is the incoming request.
		strings.NewReader("token=mF_9.B5f-4.1JqM&token_type_hint=access_token"))
	r.Header.Set("Content-Type", introspection.FormContentType)

	req, err := introspection.ParseRequest(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(req.Token, req.TokenTypeHint)
	// Output: mF_9.B5f-4.1JqM access_token
}

// Read a service-specific extension member that has no typed field.
func ExampleResponse_GetExtra() {
	var resp introspection.Response
	if err := json.Unmarshal([]byte(`{"active":true,"amr":["pwd","otp"]}`), &resp); err != nil {
		log.Fatal(err)
	}
	var amr []string
	present, err := resp.GetExtra("amr", &amr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(present, amr)
	// Output: true [pwd otp]
}

// Validate that a response describes a token usable right now, with a fixed
// clock for reproducibility.
func ExampleResponse_Validate() {
	now := time.Unix(1_700_000_000, 0)
	resp := introspection.Response{
		Active: true,
		Expiry: introspection.NewNumericDate(now.Add(-time.Minute)), // expired a minute ago
	}
	err := resp.Validate(introspection.WithClock(func() time.Time { return now }))
	fmt.Println(err)
	// Output: introspection: token has expired
}
