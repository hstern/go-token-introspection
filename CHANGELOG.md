# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-20

Initial release: a standard-library-only implementation of RFC 7662 OAuth 2.0
Token Introspection.

### Added

- Typed `Request` (token + token_type_hint) with form encode/decode, and
  `Response` covering the §2.2 members, with `Audience` (string-or-array),
  `NumericDate` (integer seconds), and byte-stable open-extension passthrough.
- Resource-server `Client` with `Introspect`, `WithHTTPClient` / `WithBasicAuth`
  options, and a typed error model (`ErrUnauthorized`, `ErrUnexpectedStatus`,
  `ErrInvalidResponse`, `*HTTPError`). An inactive token is a normal answer, not
  an error (§2.3).
- Authorization-server responder helpers `ParseRequest` and `WriteResponse`
  (no HTTP handler, by design), with `*ValidationError`.
- Typed extension accessors `Response.GetExtra` / `SetExtra`, and opt-in
  `Response.Validate` (`WithClock`, `WithLeeway`) reporting `ErrTokenInactive` /
  `ErrTokenExpired` / `ErrTokenNotYetValid`.
- Spec-derived conformance fixtures driven through both roles.

[0.1.0]: https://github.com/hstern/go-token-introspection/releases/tag/v0.1.0
