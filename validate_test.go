// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"testing"
	"time"
)

// fixedClock returns a clock that always reports t.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestValidate(t *testing.T) {
	now := time.Unix(1_000_000_000, 0).UTC()
	at := func(d time.Duration) *NumericDate { return NewNumericDate(now.Add(d)) }

	cases := []struct {
		name string
		resp Response
		opts []ValidateOption
		want error
	}{
		{"active no times", Response{Active: true}, nil, nil},
		{"inactive", Response{Active: false}, nil, ErrTokenInactive},
		{"valid window", Response{Active: true, Expiry: at(time.Hour), NotBefore: at(-time.Hour)}, nil, nil},
		{"expired", Response{Active: true, Expiry: at(-time.Minute)}, nil, ErrTokenExpired},
		{"not yet valid", Response{Active: true, NotBefore: at(time.Minute)}, nil, ErrTokenNotYetValid},
		{"expired within leeway", Response{Active: true, Expiry: at(-30 * time.Second)}, []ValidateOption{WithLeeway(time.Minute)}, nil},
		{"nbf within leeway", Response{Active: true, NotBefore: at(30 * time.Second)}, []ValidateOption{WithLeeway(time.Minute)}, nil},
		{"inactive beats expiry", Response{Active: false, Expiry: at(time.Hour)}, nil, ErrTokenInactive},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := append([]ValidateOption{WithClock(fixedClock(now))}, c.opts...)
			if err := c.resp.Validate(opts...); !errors.Is(err, c.want) {
				t.Fatalf("Validate() = %v, want %v", err, c.want)
			}
		})
	}
}

func TestValidateDefaultClock(t *testing.T) {
	// With the default (real) clock, a token expiring an hour ago is expired and
	// one expiring an hour from now is valid.
	past := Response{Active: true, Expiry: NewNumericDate(time.Now().Add(-time.Hour))}
	if err := past.Validate(); !errors.Is(err, ErrTokenExpired) {
		t.Errorf("past token: %v, want ErrTokenExpired", err)
	}
	future := Response{Active: true, Expiry: NewNumericDate(time.Now().Add(time.Hour))}
	if err := future.Validate(); err != nil {
		t.Errorf("future token: %v, want nil", err)
	}
}
