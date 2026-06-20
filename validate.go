// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"time"
)

// Validity errors reported by Response.Validate. Match them with errors.Is.
var (
	// ErrTokenInactive is reported when the response's active member is false —
	// the authorization server considers the token not currently active (§2.2).
	ErrTokenInactive = errors.New("introspection: token is not active")

	// ErrTokenExpired is reported when exp is present and in the past (§2.2).
	ErrTokenExpired = errors.New("introspection: token has expired")

	// ErrTokenNotYetValid is reported when nbf is present and in the future
	// (§2.2).
	ErrTokenNotYetValid = errors.New("introspection: token is not yet valid")
)

type validateConfig struct {
	clock  func() time.Time
	leeway time.Duration
}

// ValidateOption configures Response.Validate.
type ValidateOption func(*validateConfig)

// WithClock sets the clock Validate reads the current time from. The default is
// time.Now. Pass a fixed clock in tests for deterministic results.
func WithClock(clock func() time.Time) ValidateOption {
	return func(c *validateConfig) {
		if clock != nil {
			c.clock = clock
		}
	}
}

// WithLeeway allows up to d of clock skew when checking the exp and nbf time
// bounds. The default is zero.
func WithLeeway(d time.Duration) ValidateOption {
	return func(c *validateConfig) { c.leeway = d }
}

// Validate is an opt-in check that the response describes a token usable right
// now: active is true, and the exp/nbf time bounds (when present) place the
// current time within the token's validity window, allowing for any configured
// leeway. It returns the first failing condition as ErrTokenInactive,
// ErrTokenExpired, or ErrTokenNotYetValid, or nil if the token is usable.
//
// Validate is deliberately separate from decoding: the codec is liberal
// (RFC 7662 leaves the active determination to the server), and a consumer that
// only needs the active flag can read Active directly without calling Validate.
func (r *Response) Validate(opts ...ValidateOption) error {
	cfg := validateConfig{clock: time.Now}
	for _, opt := range opts {
		opt(&cfg)
	}
	now := cfg.clock()

	if !r.Active {
		return ErrTokenInactive
	}
	if r.Expiry != nil && r.Expiry.Before(now.Add(-cfg.leeway)) {
		return ErrTokenExpired
	}
	if r.NotBefore != nil && now.Add(cfg.leeway).Before(r.NotBefore.Time) {
		return ErrTokenNotYetValid
	}
	return nil
}
