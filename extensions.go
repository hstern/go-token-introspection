// Copyright 2026 The go-token-introspection Authors
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"encoding/json"
	"fmt"
)

// GetExtra unmarshals the service-specific extension member named name (§2.2)
// into v, which must be a non-nil pointer. It reports whether the member was
// present; a missing member is not an error (present == false, err == nil).
//
// Only members not captured by a typed field land in Extra, so GetExtra is the
// way to read extension members byte-for-byte as the server sent them.
func (r *Response) GetExtra(name string, v any) (present bool, err error) {
	raw, ok := r.Extra[name]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return true, fmt.Errorf("introspection: extension %q: %w", name, err)
	}
	return true, nil
}

// SetExtra marshals v and stores it as the extension member named name (§2.2).
// It returns an error if name collides with a member that has its own typed
// field — set those through the field instead — or if v cannot be marshalled.
func (r *Response) SetExtra(name string, v any) error {
	if _, typed := knownResponseMembers[name]; typed {
		return fmt.Errorf("introspection: %q has a typed field; set it directly", name)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("introspection: extension %q: %w", name, err)
	}
	if r.Extra == nil {
		r.Extra = make(map[string]json.RawMessage, 1)
	}
	r.Extra[name] = raw
	return nil
}
