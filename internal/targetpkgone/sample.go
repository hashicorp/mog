// Copyright IBM Corp. 2022, 2026
// SPDX-License-Identifier: MPL-2.0

package targetpkgone

type TheSample struct {
	BoolField       bool
	StringPtrField  *string
	IntField        int
	ExtraField      string
	unexportedField bool
}
