// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package targetpkgone

type TheSample struct {
	BoolField       bool
	StringPtrField  *string
	IntField        int
	ExtraField      string
	unexportedField bool
}
