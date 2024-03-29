// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStructAnnotation_Full(t *testing.T) {
	comment := `// SourceStruct does something. This comment
// spans multiple lines.
//
// mog annotation:
//
// target=github.com/hashicorp/consul/structs.Node
// output=node.gen.go
// name=Structs
// ignore-fields=RaftIndex,HiddenField,TheThirdOne
// func-from=convNodeToStructs
// func-to=convStructsToNode
`
	cfg, err := parseStructAnnotation("SourceStruct", newCommentList(comment))
	require.NoError(t, err)

	expected := structConfig{
		Source:           "SourceStruct",
		Target:           target{"github.com/hashicorp/consul/structs", "Node"},
		Output:           "node.gen.go",
		FuncNameFragment: "Structs",
		IgnoreFields:     newStringSetFromSlice([]string{"RaftIndex", "HiddenField", "TheThirdOne"}),
		FuncFrom:         "convNodeToStructs",
		FuncTo:           "convStructsToNode",
	}
	require.Equal(t, expected, cfg)
}

func newCommentList(s string) []*ast.Comment {
	var c []*ast.Comment
	for _, line := range strings.Split(s, "\n") {
		c = append(c, &ast.Comment{Text: line})
	}
	return c
}

func TestParseStructAnnotation(t *testing.T) {
	type testCase struct {
		name     string
		comment  string
		expected structConfig
	}
	fn := func(t *testing.T, tc testCase) {
		cfg, err := parseStructAnnotation("SourceStruct", newCommentList(tc.comment))
		require.NoError(t, err)
		require.Equal(t, tc.expected, cfg)
	}

	var testCases = []testCase{
		{
			name:     "annotation on last line",
			comment:  "// This is a bad comment\n// mog annotation:",
			expected: structConfig{Source: "SourceStruct"},
		},
		{
			name: "no extra newlines",
			comment: `// SourceStruct does a thing
// mog annotation:
// target=Foo name=Other`,
			expected: structConfig{
				Source:           "SourceStruct",
				Target:           target{Struct: "Foo"},
				FuncNameFragment: "Other",
			},
		},
		{
			name: "no leading comment",
			comment: `// mog annotation:
// target=Foo name=Other`,
			expected: structConfig{
				Source:           "SourceStruct",
				Target:           target{Struct: "Foo"},
				FuncNameFragment: "Other",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn(t, tc)
		})
	}
}

func TestParseStructAnnotation_Errors(t *testing.T) {
	type testCase struct {
		name    string
		comment string
		err     string
	}
	fn := func(t *testing.T, tc testCase) {
		_, err := parseStructAnnotation("SourceStruct", newCommentList(tc.comment))
		require.Error(t, err)
		require.Contains(t, err.Error(), tc.err)
	}

	var testCases = []testCase{
		{
			name:    "missing annotation identifier",
			comment: "// super-size=thing",
			err:     "missing struct annotation",
		},
		{
			name:    "unsupported annotation key",
			comment: "// mog annotation:\n// super-size=thing",
			err:     "invalid annotation key super-size in term 'super-size=thing'",
		},
		{
			name:    "invalid term, missing =",
			comment: "// mog annotation:\n// target",
			err:     "invalid term 'target' in annotation, expected only one =",
		},
		{
			name:    "invalid term, too many =",
			comment: "// mog annotation:\n// target=Foo=Thing",
			err:     "invalid term 'target=Foo=Thing' in annotation, expected only one =",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn(t, tc)
		})
	}
}

func TestParseFieldAnnotation_Full(t *testing.T) {
	comment := `// Some field
// mog: target=ID pointer=zero-is-nil,copy func-from=string func-to=types.ID`

	field := &ast.Field{
		Doc:   &ast.CommentGroup{List: newCommentList(comment)},
		Names: []*ast.Ident{{Name: "Some"}},
		Type:  &ast.Ident{Name: "string"},
	}
	cfg, err := parseFieldAnnotation(field)
	require.NoError(t, err)
	expected := fieldConfig{
		SourceName: "Some",
		SourceExpr: field.Type,
		TargetName: "ID",
		FuncFrom:   "string",
		FuncTo:     "types.ID",
		// TODO: pointer settings
	}
	require.Equal(t, expected, cfg)
}

// TODO: no leading comment
// TODO: extra newlines before annotation
// TODO: extra lines after annotation
// TODO: no docstring
// TODO: anonymous field?
// TODO: invalid term (too many =, missing =)
// TODO: invalid key in term
