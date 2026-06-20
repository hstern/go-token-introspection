# go-token-introspection

[![Go Reference](https://pkg.go.dev/badge/github.com/hstern/go-token-introspection.svg)](https://pkg.go.dev/github.com/hstern/go-token-introspection)

A typed client and request/response model for **RFC 7662 — OAuth 2.0 Token
Introspection** (Proposed Standard, 2015-10).
Spec: <https://www.rfc-editor.org/rfc/rfc7662.html>

A resource server uses token introspection to ask an authorization server
whether an opaque access token is active and to read its metadata (`scope`,
`client_id`, `exp`, `sub`, `aud`, …). This library provides the typed wire
shapes, a small HTTP client for the resource-server side, and à-la-carte parse
and write helpers for the authorization-server side.

Standard library only. No JOSE, no JWT, no framework glue.

## Install

```bash
go get github.com/hstern/go-token-introspection
```

Requires Go 1.26+.

## Resource-server side: the client

```go
c := introspection.NewClient("https://as.example.com/introspect",
	introspection.WithBasicAuth("resource-server", "secret"))

resp, err := c.Introspect(ctx, &introspection.Request{Token: tok})
if err != nil {
	// transport failure, a non-200 status, or an undecodable body
	return err
}
if !resp.Active {
	// the token is not active — a normal answer, not an error
	return errUnauthorized
}
// resp.Scope, resp.ClientID, resp.Expiry, resp.Audience, ...
```

An inactive or unknown token comes back as a `Response` with `Active == false`
and a **nil error** (RFC 7662 §2.3) — it is a normal answer, not a failure.
Errors are reserved for the request not completing:

```go
switch {
case errors.Is(err, introspection.ErrUnauthorized):
	// the AS rejected this resource server's own client authentication (401)
case errors.Is(err, introspection.ErrUnexpectedStatus):
	var he *introspection.HTTPError
	errors.As(err, &he) // he.StatusCode has the exact code
case errors.Is(err, introspection.ErrInvalidResponse):
	// a 200 body that would not decode
}
```

Client authentication beyond HTTP Basic — mTLS, private-key-JWT, secret-in-body
— along with TLS and timeouts are transport concerns; configure them on the
`*http.Client` passed to `WithHTTPClient`.

## Authorization-server side: the responder helpers

The library ships à-la-carte helpers, not an `http.Handler` — routing, caller
authentication, and token lookup stay in the endpoint you already run:

```go
func handleIntrospect(w http.ResponseWriter, r *http.Request) {
	// ... authenticate the caller first (RFC 7662 §2.1) ...

	req, err := introspection.ParseRequest(r)
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	resp := lookUpToken(req.Token) // your policy; build a *introspection.Response
	_ = introspection.WriteResponse(w, resp)
}
```

`ParseRequest` enforces the form content type and the required `token`
parameter (returning a `*ValidationError` otherwise); `WriteResponse` emits the
JSON body with `active` always present.

## Extension members

Service-specific members beyond the twelve RFC 7662 names are preserved
byte-for-byte and read or written with typed accessors:

```go
var amr []string
present, err := resp.GetExtra("amr", &amr)

err = resp.SetExtra("acr", "urn:mace:incommon:iap:silver")
```

## Validation

Decoding is liberal. For a consumer that wants one call to answer "is this token
usable right now?", `Validate` checks `active` and the `exp`/`nbf` window:

```go
if err := resp.Validate(introspection.WithLeeway(30 * time.Second)); err != nil {
	// introspection.ErrTokenInactive, ErrTokenExpired, or ErrTokenNotYetValid
}
```

## Scope

In scope: the typed `Request`/`Response`, the resource-server `Client`, the
à-la-carte responder helpers, extension passthrough, and opt-in validation.

Out of scope, by design:

- **Client authentication schemes** other than HTTP Basic — transport-injected.
- **The authorization decision.** Surfacing `aud`/`scope` is in scope; deciding
  whether an active token is *for this resource server* is your policy.
- **JWT/JWS verification.** Introspection is the opaque-token alternative to
  local JWT validation; verifying a signed token is a JOSE concern.

## Versioning

Semantic Versioning, tracked independently of the spec. The current series is
`v0.x`: the public API may still change before `v1.0.0`. The targeted spec
version is exposed as `introspection.SpecVersion`.

## License

Apache-2.0 — see [LICENSE](LICENSE).
