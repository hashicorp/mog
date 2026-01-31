// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"fmt"
	"go/types"
)

// assignmentKind is implemented by all general types of mapping operations
// between types.
type assignmentKind interface {
	isAssignmentKind()
	fmt.Stringer
}

type specialAssignmentAttributes struct {
	// Convenience helpers
	ProtobufTimestamp bool
	ProtobufDuration  bool
	ProtobufDirection Direction
}

func (s specialAssignmentAttributes) ReverseDirection() specialAssignmentAttributes {
	s2 := s
	s2.ProtobufDirection = s.ProtobufDirection.Reverse()
	return s2
}

func (s specialAssignmentAttributes) IsZero() bool {
	return s == specialAssignmentAttributes{}
}

func (s specialAssignmentAttributes) GoString() string {
	return s.String()
}

func (s specialAssignmentAttributes) String() string {
	switch {
	case s.ProtobufTimestamp:
		return "protobuf " + s.ProtobufDirection.String() + " timestamp"
	case s.ProtobufDuration:
		return "protobuf " + s.ProtobufDirection.String() + " duration"
	default:
		return ""
	}
}

// singleAssignmentKind is a mapping operation between two fields that
// ultimately are:
//
//   - basic
//   - named structs
//   - pointers to either of the above
type singleAssignmentKind struct {
	// Left is the original type of the LHS of the assignment.
	Left types.Type

	// Right is the original type of the RHS of the assignment.
	Right types.Type

	// Direct implies that no conversion is needed and direct assignment should
	// occur.
	Direct bool

	// Convert implies that a simple type conversion is required.
	Convert bool

	Special specialAssignmentAttributes
}

var _ assignmentKind = (*singleAssignmentKind)(nil)

func (o *singleAssignmentKind) isAssignmentKind() {}
func (o *singleAssignmentKind) String() string {
	s := fmt.Sprintf("%s := %s", debugPrintType(o.Left), debugPrintType(o.Right))
	if o.Direct {
		s += " (direct)"
	}
	if o.Convert {
		s += " (convert)"
	}
	if !o.Special.IsZero() {
		s += " (" + o.Special.String() + ")"
	}
	return s
}

// sliceAssignmentKind is a mapping operation between two fields that are
// slice-ish and have elements that would satisfy singleAssignmentKind
type sliceAssignmentKind struct {
	// Left is the original type of the LHS of the assignment. Should be
	// slice-ish.
	Left types.Type

	// Right is the original type of the RHS of the assignment. Should be
	// slice-ish.
	Right types.Type

	// LeftElem is the original type of the elements of the LHS of the
	// assignment.
	LeftElem types.Type

	// RightElem is the original type of the elements of the LHS of the
	// assignment.
	RightElem types.Type

	// ElemDirect implies that no conversion is needed and direct assignment
	// should occur for elements of the slice.
	ElemDirect bool

	// ElemConvert implies that a simple type conversion is required for
	// elements of the slice.
	ElemConvert bool

	ElemSpecial specialAssignmentAttributes
}

var _ assignmentKind = (*sliceAssignmentKind)(nil)

func (o *sliceAssignmentKind) isAssignmentKind() {}
func (o *sliceAssignmentKind) String() string {
	s := fmt.Sprintf("%s[%s] := %s[%s]",
		debugPrintType(o.Left),
		debugPrintType(o.LeftElem),
		debugPrintType(o.Right),
		debugPrintType(o.RightElem),
	)
	if o.ElemDirect {
		s += " (direct)"
	}
	if o.ElemConvert {
		s += " (convert)"
	}
	if !o.ElemSpecial.IsZero() {
		s += " (" + o.ElemSpecial.String() + ")"
	}
	return s
}

// mapAssignmentKind is a mapping operation between two fields that are map-ish
// and have value elements that would satisfy singleAssignmentKind and key
// elements that are directly assignable.
type mapAssignmentKind struct {
	// Left is the original type of the LHS of the assignment. Should be
	// map-ish.
	Left types.Type

	// Right is the original type of the RHS of the assignment. Should be
	// map-ish.
	Right types.Type

	// LeftKey is the original type of the keys of the LHS of the
	// assignment.
	LeftKey types.Type

	// RightKey is the original type of the keys of the LHS of the
	// assignment.
	RightKey types.Type

	// LeftElem is the original type of the elements of the LHS of the
	// assignment.
	LeftElem types.Type

	// RightElem is the original type of the elements of the LHS of the
	// assignment.
	RightElem types.Type

	// ElemDirect implies that no conversion is needed and direct assignment
	// should occur for elements of the map.
	ElemDirect bool

	// ElemConvert implies that a simple type conversion is required for
	// elements of the map.
	ElemConvert bool

	ElemSpecial specialAssignmentAttributes
}

var _ assignmentKind = (*mapAssignmentKind)(nil)

func (o *mapAssignmentKind) isAssignmentKind() {}
func (o *mapAssignmentKind) String() string {
	s := fmt.Sprintf("%s<%s,%s> := %s<%s,%s>",
		debugPrintType(o.Left),
		debugPrintType(o.LeftKey),
		debugPrintType(o.LeftElem),
		debugPrintType(o.Right),
		debugPrintType(o.RightKey),
		debugPrintType(o.RightElem),
	)
	if o.ElemDirect {
		s += " (direct)"
	}
	if o.ElemConvert {
		s += " (convert)"
	}
	if !o.ElemSpecial.IsZero() {
		s += " (" + o.ElemSpecial.String() + ")"
	}
	return s
}

func convertibleButNotIdentical(typ, typeDecode types.Type) bool {
	// Only allow this for basic and named.

	switch typeDecode.(type) {
	case *types.Basic, *types.Named:
	default:
		return false
	}

	if types.Identical(typ, typeDecode) {
		return false
	}

	if types.ConvertibleTo(typ, typeDecode) && types.ConvertibleTo(typeDecode, typ) {
		return true
	}
	return false
}

func isAnyProtobufTimeOrDuration(types ...types.Type) bool {
	for _, typ := range types {
		switch {
		case isProtobufDuration(typ), isProtobufTimestamp(typ):
			return true
		}
	}
	return false
}

func isProtobufDuration(typ types.Type) bool {
	pt, ok := typ.(*types.Pointer)
	if !ok {
		return false
	}
	if nt, ok := pt.Elem().(*types.Named); ok {
		return nt.String() == "github.com/golang/protobuf/ptypes/duration.Duration"
	}
	return false
}

func isProtobufTimestamp(typ types.Type) bool {
	pt, ok := typ.(*types.Pointer)
	if !ok {
		return false
	}
	if nt, ok := pt.Elem().(*types.Named); ok {
		return nt.String() == "github.com/golang/protobuf/ptypes/timestamp.Timestamp"
	}
	return false
}

func isGoDuration(typ types.Type) bool {
	if nt, ok := typ.(*types.Named); ok {
		return nt.String() == "time.Duration"
	}
	return false
}

func isGoTime(typ types.Type) bool {
	if nt, ok := typ.(*types.Named); ok {
		return nt.String() == "time.Time"
	}
	return false
}

var errAssignmentNotSupported = errors.New("assignment not supported")

// computeAssignment attempts to determine how to assign something of the
// rightType to something of the leftType.
//
// If this is not possible, or not currently supported
// errAssignmentNotSupported is returned.
func computeAssignment(leftType, rightType types.Type, imports *imports) (assignmentKind, error) {
	// First check if the types are naturally directly assignable. Only allow
	// type pairs that are symmetrically assignable for simplicity.
	if types.AssignableTo(rightType, leftType) {
		if !types.AssignableTo(leftType, rightType) {
			return nil, errAssignmentNotSupported
		}
		return &singleAssignmentKind{
			Left:   leftType,
			Right:  rightType,
			Direct: true,
		}, nil
	}

	// We don't really care about type aliases or pointerness here, so peel
	// those off first to simplify the space we have to consider below.
	leftTypeDecode, leftOk := decodeType(leftType)
	rightTypeDecode, rightOk := decodeType(rightType)
	if !leftOk || !rightOk {
		return nil, errAssignmentNotSupported
	}

	if isAnyProtobufTimeOrDuration(leftType, rightType) {
		// Note: we only do conversions for non-pointer stdlib to pointer
		// protobufs with no aliasing.

		kind := &singleAssignmentKind{
			Left:  leftType,
			Right: rightType,
		}
		switch {
		case isProtobufTimestamp(leftType) && isGoTime(rightType):
			kind.Special.ProtobufTimestamp = true
			kind.Special.ProtobufDirection = DirTo
		case isProtobufDuration(leftType) && isGoDuration(rightType):
			kind.Special.ProtobufDuration = true
			kind.Special.ProtobufDirection = DirTo
		case isGoTime(leftType) && isProtobufTimestamp(rightType):
			kind.Special.ProtobufTimestamp = true
			kind.Special.ProtobufDirection = DirFrom
		case isGoDuration(leftType) && isProtobufDuration(rightType):
			kind.Special.ProtobufDuration = true
			kind.Special.ProtobufDirection = DirFrom
		default:
			// will require a helper function
			return nil, fmt.Errorf("one struct field is a protobuf time or duration and requires a user func")
		}

		imports.Add("", "github.com/golang/protobuf/ptypes")

		return kind, nil
	}

	if convertibleButNotIdentical(rightType, rightTypeDecode) ||
		convertibleButNotIdentical(leftType, leftTypeDecode) {

		return &singleAssignmentKind{
			Left:    leftType,
			Right:   rightType,
			Convert: true,
		}, nil
	}

	switch left := leftTypeDecode.(type) {
	case *types.Basic:
		// basic can only assign to basic
		_, ok := rightTypeDecode.(*types.Basic)
		if !ok {
			return nil, errAssignmentNotSupported
		}
		return &singleAssignmentKind{
			Left:   leftType,
			Right:  rightType,
			Direct: true,
		}, nil
	case *types.Named:
		// named can only assign to named
		_, ok := rightTypeDecode.(*types.Named)
		if !ok {
			return nil, errAssignmentNotSupported
		}
		return &singleAssignmentKind{
			Left:  leftType,
			Right: rightType,
		}, nil
	case *types.Slice:
		// slices can only assign to slices
		right, ok := rightTypeDecode.(*types.Slice)
		if !ok {
			return nil, errAssignmentNotSupported
		}

		// the elements have to be assignable
		rawOp, err := computeAssignment(left.Elem(), right.Elem(), imports)
		if err != nil {
			return nil, err
		}

		op, ok := rawOp.(*singleAssignmentKind)
		if !ok {
			return nil, errAssignmentNotSupported
		}

		return &sliceAssignmentKind{
			Left:        leftType,
			LeftElem:    left.Elem(),
			Right:       rightType,
			RightElem:   right.Elem(),
			ElemDirect:  op.Direct,
			ElemConvert: op.Convert,
			ElemSpecial: op.Special,
		}, nil
	case *types.Map:
		right, ok := rightTypeDecode.(*types.Map)
		if !ok {
			return nil, errAssignmentNotSupported
		}

		rawKeyOp, err := computeAssignment(left.Key(), right.Key(), imports)
		if err != nil {
			return nil, err
		}

		// the map keys have to be directly assignable
		keyOp, ok := rawKeyOp.(*singleAssignmentKind)
		if !ok {
			return nil, errAssignmentNotSupported
		}
		if !keyOp.Direct {
			return nil, errAssignmentNotSupported
		}

		// the map values have to be assignable
		rawOp, err := computeAssignment(left.Elem(), right.Elem(), imports)
		if err != nil {
			return nil, err
		}

		op, ok := rawOp.(*singleAssignmentKind)
		if !ok {
			return nil, errAssignmentNotSupported
		}

		return &mapAssignmentKind{
			Left:        leftType,
			LeftKey:     left.Key(),
			LeftElem:    left.Elem(),
			Right:       rightType,
			RightKey:    right.Key(),
			RightElem:   right.Elem(),
			ElemDirect:  op.Direct,
			ElemConvert: op.Convert,
			ElemSpecial: op.Special,
		}, nil
	}

	return nil, errAssignmentNotSupported
}
