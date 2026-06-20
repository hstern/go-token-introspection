// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Response is an RFC 7662 §2.2 introspection response.
//
// Active is the only REQUIRED member; every other field is optional and is
// omitted from the wire when unset. Service-specific extension members (§2.2)
// are preserved verbatim in Extra for byte-stable round-trips.
//
// Decoding is liberal (Postel's law): UnmarshalJSON accepts whatever the wire
// provides. Strict checks are opt-in via Validate.
type Response struct {
	// Active indicates whether the token is currently active (§2.2). REQUIRED.
	// A well-formed, authorized query for an inactive or unknown token returns
	// Active == false — that is a normal response, not an error (§2.3).
	Active bool `json:"active"`

	// Scope is a space-separated list of scopes (§2.2, RFC 6749 §3.3). Stored
	// verbatim; use Scopes to split it.
	Scope string `json:"scope,omitempty"`

	// ClientID is the client the token was issued to (§2.2).
	ClientID string `json:"client_id,omitempty"`

	// Username is a human-readable identifier for the resource owner (§2.2). It
	// is not necessarily the same as Subject.
	Username string `json:"username,omitempty"`

	// TokenType is the type of the token, e.g. "Bearer" (§2.2, RFC 6749 §7.1).
	TokenType string `json:"token_type,omitempty"`

	// Expiry (exp), IssuedAt (iat), and NotBefore (nbf) are integer timestamps,
	// seconds since the Unix epoch (§2.2).
	Expiry    *NumericDate `json:"exp,omitempty"`
	IssuedAt  *NumericDate `json:"iat,omitempty"`
	NotBefore *NumericDate `json:"nbf,omitempty"`

	// Subject is the subject of the token (§2.2); usually a machine-readable
	// identifier of the resource owner.
	Subject string `json:"sub,omitempty"`

	// Audience is the service-specific audience of the token (§2.2): a single
	// string identifier or a list of them.
	Audience Audience `json:"aud,omitempty"`

	// Issuer is the issuer of the token (§2.2).
	Issuer string `json:"iss,omitempty"`

	// JWTID is a unique identifier for the token (§2.2, jti).
	JWTID string `json:"jti,omitempty"`

	// Extra holds any response members not captured by the typed fields above —
	// service-specific extensions and future registrations (§2.2). Values are
	// kept as raw JSON for byte-stable round-trips and zero-cost pass-through.
	Extra map[string]json.RawMessage `json:"-"`
}

// Scopes splits the space-delimited Scope into its individual scope tokens
// (RFC 6749 §3.3). An empty Scope yields nil.
func (r *Response) Scopes() []string {
	return strings.Fields(r.Scope)
}

// knownResponseMembers are the JSON keys mapped to typed Response fields.
// Anything else decoded from a response object lands in Extra.
var knownResponseMembers = map[string]struct{}{
	"active": {}, "scope": {}, "client_id": {}, "username": {},
	"token_type": {}, "exp": {}, "iat": {}, "nbf": {}, "sub": {},
	"aud": {}, "iss": {}, "jti": {},
}

// UnmarshalJSON decodes the typed members and routes every other member of the
// JSON object into Extra.
func (r *Response) UnmarshalJSON(data []byte) error {
	type alias Response
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = Response(a)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	for k := range knownResponseMembers {
		delete(all, k)
	}
	if len(all) > 0 {
		r.Extra = all
	}
	return nil
}

// MarshalJSON serializes the typed members and merges Extra back in. Typed
// members win on key collision. Output is byte-stable: with no extension members
// the typed members serialize in their declared order; with extensions the whole
// object serializes in encoding/json's sorted-key order. Either way a given
// Response value always marshals to the same bytes.
func (r Response) MarshalJSON() ([]byte, error) {
	type alias Response
	known, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	if len(r.Extra) == 0 {
		return known, nil
	}

	merged := make(map[string]json.RawMessage, len(r.Extra)+len(knownResponseMembers))
	if err := json.Unmarshal(known, &merged); err != nil {
		return nil, err
	}
	for k, v := range r.Extra {
		if _, taken := merged[k]; taken {
			continue
		}
		merged[k] = v
	}
	return json.Marshal(merged)
}

// Audience is a token audience: one or more service-specific string identifiers
// (§2.2). It decodes from either a single JSON string or an array of strings,
// and marshals back to a single string when it holds exactly one element.
type Audience []string

// Contains reports whether aud is present in the audience.
func (a Audience) Contains(aud string) bool {
	return slices.Contains(a, aud)
}

// UnmarshalJSON accepts either a single string or an array of strings.
func (a *Audience) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = Audience{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

// MarshalJSON emits a bare string for a single-element audience and a JSON
// array otherwise.
func (a Audience) MarshalJSON() ([]byte, error) {
	if len(a) == 1 {
		return json.Marshal(a[0])
	}
	return json.Marshal([]string(a))
}

// NumericDate is an RFC 7662 timestamp: an integer number of seconds since the
// Unix epoch (§2.2). It marshals to integer seconds and decodes liberally,
// tolerating a fractional part.
type NumericDate struct {
	time.Time
}

// NewNumericDate returns a *NumericDate for t.
func NewNumericDate(t time.Time) *NumericDate {
	return &NumericDate{t}
}

// MarshalJSON emits the timestamp as integer seconds since the Unix epoch.
func (n NumericDate) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, n.Unix(), 10), nil
}

// UnmarshalJSON parses a JSON number of seconds since the Unix epoch, tolerating
// a fractional part.
func (n *NumericDate) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	sec, frac := math.Modf(f)
	n.Time = time.Unix(int64(sec), int64(math.Round(frac*1e9))).UTC()
	return nil
}
