// Copyright IBM Corp. 2022, 2026
// SPDX-License-Identifier: MPL-2.0

package targetpkgtwo

type Lamp struct {
	Brand   string
	Sockets uint8
}

type Flood struct {
	StructIsAlsoAField bool
}

type StructIsAlsoAField struct {
	ID Identifier
}

type Identifier struct {
	Name      string
	Namespace string
}
