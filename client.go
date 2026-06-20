// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls an RFC 7662 introspection endpoint on behalf of a protected
// resource. It is safe for concurrent use; the zero value is not usable — build
// one with NewClient.
type Client struct {
	endpoint   string
	httpClient *http.Client
	basicAuth  *basicAuth
}

type basicAuth struct {
	clientID     string
	clientSecret string
}

// ClientOption configures a Client in NewClient.
type ClientOption func(*Client)

// NewClient returns a Client that introspects tokens at endpoint, a fully
// qualified introspection endpoint URL (RFC 7662 §2). By default it uses
// http.DefaultClient and sends no authentication; override with WithHTTPClient
// and WithBasicAuth.
func NewClient(endpoint string, opts ...ClientOption) *Client {
	c := &Client{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithHTTPClient sets the HTTP client used for introspection requests. This is
// the injection point for transport-level concerns the library deliberately
// leaves to the caller: TLS configuration, timeouts, and any client
// authentication scheme other than HTTP Basic (RFC 7662 §2.1). A nil client is
// ignored.
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithBasicAuth authenticates introspection requests with HTTP Basic, sending
// clientID and clientSecret as the credentials (RFC 7662 §2.1, RFC 6749 §2.3.1)
// — the one client-authentication scheme this library implements directly.
// Every other scheme (private-key-JWT, mTLS, secret-in-body) belongs on the
// transport supplied via WithHTTPClient.
func WithBasicAuth(clientID, clientSecret string) ClientOption {
	return func(c *Client) {
		c.basicAuth = &basicAuth{clientID: clientID, clientSecret: clientSecret}
	}
}

// Errors returned by Introspect. Use errors.Is to test for them; an *HTTPError
// (via errors.As) carries the exact status code.
var (
	// ErrUnauthorized is reported when the endpoint rejects the caller's own
	// client authentication with HTTP 401 (RFC 7662 §2.3). It is wrapped by the
	// *HTTPError returned for a 401 response.
	ErrUnauthorized = errors.New("introspection: endpoint rejected client authentication")

	// ErrUnexpectedStatus is reported for any non-200, non-401 HTTP status. It
	// is wrapped by the *HTTPError returned for such a response.
	ErrUnexpectedStatus = errors.New("introspection: unexpected response status")

	// ErrInvalidResponse is reported when a 200 response body is not valid JSON
	// or cannot be decoded into a Response.
	ErrInvalidResponse = errors.New("introspection: malformed response body")
)

// maxResponseBytes bounds how much of a response body Introspect will read, so
// a misbehaving or hostile endpoint cannot exhaust memory. Introspection
// responses are small; 1 MiB is comfortably above any legitimate payload.
const maxResponseBytes = 1 << 20

// HTTPError reports an introspection response whose status was not 200 OK. It
// wraps ErrUnauthorized for a 401 and ErrUnexpectedStatus otherwise, so callers
// can match either the specific or the general case with errors.Is, or read the
// code with errors.As.
type HTTPError struct {
	StatusCode int    // e.g. 401
	Status     string // e.g. "401 Unauthorized"
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("introspection: unexpected response status %q", e.Status)
}

func (e *HTTPError) Unwrap() error {
	if e.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	return ErrUnexpectedStatus
}

// Introspect submits req to the introspection endpoint and returns the decoded
// Response.
//
// A token the authorization server reports as inactive or unknown is a normal
// result, not an error: the call returns a Response with Active == false and a
// nil error (RFC 7662 §2.2, §2.3). Errors are reserved for the request failing
// to complete: a transport failure, a non-200 status (*HTTPError, wrapping
// ErrUnauthorized or ErrUnexpectedStatus), or a body that will not decode
// (ErrInvalidResponse).
func (c *Client) Introspect(ctx context.Context, req *Request) (*Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(req.EncodeForm()))
	if err != nil {
		return nil, fmt.Errorf("introspection: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", FormContentType)
	httpReq.Header.Set("Accept", "application/json")
	if c.basicAuth != nil {
		httpReq.SetBasicAuth(c.basicAuth.clientID, c.basicAuth.clientSecret)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("introspection: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, &HTTPError{StatusCode: httpResp.StatusCode, Status: httpResp.Status}
	}

	var resp Response
	if err := json.NewDecoder(io.LimitReader(httpResp.Body, maxResponseBytes)).Decode(&resp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return &resp, nil
}
