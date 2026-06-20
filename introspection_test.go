// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import "testing"

func TestSpecVersion(t *testing.T) {
	if SpecVersion == "" {
		t.Fatal("SpecVersion must be set")
	}
}
