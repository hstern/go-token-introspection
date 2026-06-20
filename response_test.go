// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"reflect"
	"testing"
)

// figure4 is the active-token example response from RFC 7662 §2.2 (Figure 4).
const figure4 = `{
  "active": true,
  "client_id": "l238j323ds-23ij4",
  "username": "jdoe",
  "scope": "read write dolphin",
  "sub": "Z5O3upPC88QrAjx00dis",
  "aud": "https://protected.example.net/resource",
  "iss": "https://server.example.com/",
  "exp": 1419356238,
  "iat": 1419350238,
  "extension_field": "twenty-seven"
}`

func TestResponseInactiveMarshalsMinimal(t *testing.T) {
	got, err := json.Marshal(Response{Active: false})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"active":false}`; string(got) != want {
		t.Fatalf("Marshal(inactive) = %s, want %s", got, want)
	}
}

func TestResponseFigure4Decode(t *testing.T) {
	var r Response
	if err := json.Unmarshal([]byte(figure4), &r); err != nil {
		t.Fatal(err)
	}
	if !r.Active {
		t.Error("Active = false, want true")
	}
	if r.ClientID != "l238j323ds-23ij4" {
		t.Errorf("ClientID = %q", r.ClientID)
	}
	if r.Username != "jdoe" {
		t.Errorf("Username = %q", r.Username)
	}
	if got := r.Scopes(); !reflect.DeepEqual(got, []string{"read", "write", "dolphin"}) {
		t.Errorf("Scopes() = %v", got)
	}
	if r.Subject != "Z5O3upPC88QrAjx00dis" {
		t.Errorf("Subject = %q", r.Subject)
	}
	if !reflect.DeepEqual(r.Audience, Audience{"https://protected.example.net/resource"}) {
		t.Errorf("Audience = %v", r.Audience)
	}
	if r.Issuer != "https://server.example.com/" {
		t.Errorf("Issuer = %q", r.Issuer)
	}
	if r.Expiry == nil || r.Expiry.Unix() != 1419356238 {
		t.Errorf("Expiry = %v", r.Expiry)
	}
	if r.IssuedAt == nil || r.IssuedAt.Unix() != 1419350238 {
		t.Errorf("IssuedAt = %v", r.IssuedAt)
	}
	// The unknown member is preserved verbatim in Extra.
	raw, ok := r.Extra["extension_field"]
	if !ok {
		t.Fatal("extension_field missing from Extra")
	}
	if string(raw) != `"twenty-seven"` {
		t.Errorf("Extra[extension_field] = %s", raw)
	}
	// No typed member leaks into Extra.
	if _, leaked := r.Extra["active"]; leaked {
		t.Error("active leaked into Extra")
	}
}

// Re-marshalling a decoded Response is idempotent: a given value always
// marshals to the same bytes, and decoding those bytes yields an equal value.
func TestResponseRoundTripStable(t *testing.T) {
	for _, src := range []string{
		figure4,
		`{"active":false}`,
		`{"active":true,"aud":["a","b"],"exp":1419356238}`,
	} {
		var r1 Response
		if err := json.Unmarshal([]byte(src), &r1); err != nil {
			t.Fatalf("unmarshal %s: %v", src, err)
		}
		b1, err := json.Marshal(r1)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var r2 Response
		if err := json.Unmarshal(b1, &r2); err != nil {
			t.Fatalf("re-unmarshal %s: %v", b1, err)
		}
		if !reflect.DeepEqual(r1, r2) {
			t.Errorf("value drift for %s:\n r1 = %+v\n r2 = %+v", src, r1, r2)
		}
		b2, err := json.Marshal(r2)
		if err != nil {
			t.Fatalf("re-marshal: %v", err)
		}
		if string(b1) != string(b2) {
			t.Errorf("byte drift for %s:\n b1 = %s\n b2 = %s", src, b1, b2)
		}
	}
}

// A nested extension object keeps its on-the-wire key order, since Extra holds
// raw JSON bytes rather than a re-serialized map.
func TestResponseExtraPreservesNestedOrder(t *testing.T) {
	src := `{"active":true,"ext":{"z":1,"a":2,"m":3}}`
	var r Response
	if err := json.Unmarshal([]byte(src), &r); err != nil {
		t.Fatal(err)
	}
	if got := string(r.Extra["ext"]); got != `{"z":1,"a":2,"m":3}` {
		t.Errorf("Extra[ext] = %s, want verbatim order", got)
	}
}

func TestAudienceStringVsArray(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Audience
		out  string
	}{
		{"single string", `"a"`, Audience{"a"}, `"a"`},
		{"multi array", `["a","b"]`, Audience{"a", "b"}, `["a","b"]`},
		// A single-element array collapses to a bare string on re-marshal
		// (the two encodings are semantically identical for an audience).
		{"single array collapses", `["a"]`, Audience{"a"}, `"a"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var a Audience
			if err := json.Unmarshal([]byte(c.in), &a); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(a, c.want) {
				t.Fatalf("Unmarshal(%s) = %v, want %v", c.in, a, c.want)
			}
			b, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != c.out {
				t.Fatalf("Marshal = %s, want %s", b, c.out)
			}
		})
	}
}

func TestNumericDate(t *testing.T) {
	var n NumericDate
	if err := json.Unmarshal([]byte("1419356238"), &n); err != nil {
		t.Fatal(err)
	}
	if n.Unix() != 1419356238 {
		t.Fatalf("Unix() = %d", n.Unix())
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "1419356238" {
		t.Fatalf("Marshal = %s, want 1419356238", b)
	}
	// A fractional part is tolerated on decode (liberal).
	if err := json.Unmarshal([]byte("1419356238.0"), &n); err != nil {
		t.Fatalf("fractional decode: %v", err)
	}
	if n.Unix() != 1419356238 {
		t.Fatalf("fractional Unix() = %d", n.Unix())
	}
}
