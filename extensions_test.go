// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"testing"
)

func TestGetExtraPresent(t *testing.T) {
	var r Response
	if err := json.Unmarshal([]byte(`{"active":true,"extension_field":"twenty-seven"}`), &r); err != nil {
		t.Fatal(err)
	}
	var s string
	present, err := r.GetExtra("extension_field", &s)
	if err != nil {
		t.Fatal(err)
	}
	if !present || s != "twenty-seven" {
		t.Fatalf("GetExtra = (%q, present=%v)", s, present)
	}
}

func TestGetExtraMissingIsNotAnError(t *testing.T) {
	var r Response
	var s string
	present, err := r.GetExtra("nope", &s)
	if err != nil || present {
		t.Fatalf("GetExtra(missing) = (present=%v, err=%v), want (false, nil)", present, err)
	}
}

func TestGetExtraTypeMismatch(t *testing.T) {
	var r Response
	if err := json.Unmarshal([]byte(`{"active":true,"n":"not-an-int"}`), &r); err != nil {
		t.Fatal(err)
	}
	var n int
	present, err := r.GetExtra("n", &n)
	if !present || err == nil {
		t.Fatalf("GetExtra type mismatch = (present=%v, err=%v), want (true, error)", present, err)
	}
}

func TestSetExtraRoundTrip(t *testing.T) {
	r := Response{Active: true}
	if err := r.SetExtra("amr", []string{"pwd", "otp"}); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var back Response
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	var amr []string
	present, err := back.GetExtra("amr", &amr)
	if err != nil || !present {
		t.Fatalf("GetExtra after round-trip = (present=%v, err=%v)", present, err)
	}
	if len(amr) != 2 || amr[0] != "pwd" || amr[1] != "otp" {
		t.Fatalf("amr = %v", amr)
	}
}

func TestSetExtraRejectsTypedMember(t *testing.T) {
	var r Response
	if err := r.SetExtra("active", false); err == nil {
		t.Fatal("SetExtra on a typed member should error")
	}
}
