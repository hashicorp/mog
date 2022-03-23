package main

import (
	"go/types"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func BenchmarkSourceStructs(b *testing.B) {
	actual, err := loadSourceStructs("./internal/sourcepkg", "", "", packageLoadErrors)
	require.NoError(b, err)
	require.Equal(b, []string{"GroupedSample", "Sample"}, actual.StructNames())
	_, ok := actual.Structs["Sample"]
	require.True(b, ok)
	_, ok = actual.Structs["GroupedSample"]
	require.True(b, ok)

	// TODO: check the value in structs map
}

// TODO: test non-built-in types
// TODO: test types from other packages
func BenchmarkLoadTargetStructs(b *testing.B) {
	actual, err := loadTargetStructs([]string{"./internal/targetpkgone", "./internal/targetpkgtwo"}, "", "")
	assert.NilError(b, err)

	expected := map[string]targetPkg{
		"github.com/hashicorp/mog/internal/targetpkgone": {
			Structs: map[string]targetStruct{
				"TheSample": {
					Name: "TheSample",
					Fields: []*types.Var{
						newField("BoolField", types.Typ[types.Bool]),
						newField("StringPtrField", types.NewPointer(types.Typ[types.String])),
						newField("IntField", types.Typ[types.Int]),
						newField("ExtraField", types.Typ[types.String]),
					},
				},
			},
		},
		"github.com/hashicorp/mog/internal/targetpkgtwo": {
			Structs: map[string]targetStruct{
				"Lamp": {
					Name: "Lamp",
					Fields: []*types.Var{
						newField("Brand", types.Typ[types.String]),
						newField("Sockets", types.Typ[types.Uint8]),
					},
				},
				"Flood": {
					Name: "Flood",
					Fields: []*types.Var{
						newField("StructIsAlsoAField", types.Typ[types.Bool]),
					},
				},
				"StructIsAlsoAField": {
					Name: "StructIsAlsoAField",
					Fields: []*types.Var{
						newField("ID", types.NewNamed(
							types.NewTypeName(0, nil, "Identifier", nil),
							&types.Struct{},
							nil)),
					},
				},
				"Identifier": {
					Name: "Identifier",
					Fields: []*types.Var{
						newField("Name", types.Typ[types.String]),
						newField("Namespace", types.Typ[types.String]),
					},
				},
			},
		},
	}

	assert.DeepEqual(b, expected, actual, cmpTypesVar)
}

var cmpTypesVar = gocmp.Options{
	gocmp.Comparer(func(x, y *types.Var) bool {
		if x == nil || y == nil {
			return x == y
		}
		if x.Name() != y.Name() {
			return false
		}
		return gocmp.Equal(x.Type(), y.Type(), cmpTypesTypes)
	}),
}

var cmpTypesTypes = gocmp.Options{
	gocmp.AllowUnexported(types.Pointer{}, types.Basic{}),
	gocmp.Comparer(func(x, y *types.Named) bool {
		if x == nil || y == nil {
			return x == y
		}
		return x.String() != y.String()
	}),
}

func newField(name string, typ types.Type) *types.Var {
	return types.NewField(0, nil, name, typ, false)
}
