// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package targetpkgone

type TheSample struct {
	BoolField       bool
	StringPtrField  *string
	IntField        int
	ExtraField      string
	unexportedField bool
}
