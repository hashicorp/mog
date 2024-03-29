// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	stdpath "path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type sourcePkg struct {
	// Path is the absolute filesystem path to the directory which contains the
	// source package.
	Path string

	// Name of the package as it appears in the source file.
	Name string

	// TODO: buildTags string

	// Structs declared in the source package.
	Structs map[string]structDecl

	pkg *packages.Package
}

type structDecl struct {
	Doc    []*ast.Comment
	Fields []typedField
}

type typedField struct {
	Field *ast.Field
	Var   *types.Var
}

// StructNames returns a sorted slice of all the structs in the package.
func (p sourcePkg) StructNames() []string {
	names := make([]string, 0, len(p.Structs))
	for name := range p.Structs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type handlePkgLoadErr func(pkg *packages.Package) error

// loadSourceStructs scans the provided package for struct definitions that
// have mog annotations.
func loadSourceStructs(path string, tags string, handleErr handlePkgLoadErr) (sourcePkg, error) {
	p := sourcePkg{Structs: map[string]structDecl{}}
	cfg := &packages.Config{
		Mode: modeLoadAll,
	}
	if tags == "" {
		tags = os.Getenv("GOTAGS")
	}
	if tags != "" {
		cfg.BuildFlags = []string{fmt.Sprintf("-tags=%s", tags)}
	}

	var glob string
	if strings.Contains(path, "*") {
		path, glob = stdpath.Split(path)
		path = strings.TrimSuffix(path, "/") // remove trailing slash
	}

	if glob != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return p, err
		}

		cfg.ParseFile = func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			dir, err := filepath.Abs(filepath.Dir(filename))
			if err != nil {
				return nil, err
			}

			if dir == absPath {
				// matches the -source root (the one we are editing and that the glob applies to)
				name := filepath.Base(filename)
				if match, err := filepath.Match(glob, name); err != nil {
					return nil, err
				} else if !match {
					return nil, nil
				}
			}

			// This is the default ParseFile implementation.
			const mode = parser.AllErrors | parser.ParseComments
			return parser.ParseFile(fset, filename, src, mode)
		}
	}

	{
		fi, err := os.Stat(path)
		if err != nil {
			return p, err
		}
		if !fi.IsDir() {
			return p, fmt.Errorf("source argument must not be a file")
		}
	}

	pkgs, err := packages.Load(cfg, path)
	switch {
	case err != nil:
		return p, err
	case len(pkgs) == 0:
		return p, fmt.Errorf("package not found")
	case len(pkgs) > 1:
		return p, fmt.Errorf("expected only one source package")
	}

	pkg := pkgs[0]
	if err := handleErr(pkg); err != nil {
		return p, err
	}
	if len(pkg.GoFiles) < 1 {
		return p, fmt.Errorf("no Go files in the source package")
	}
	p.Path = filepath.Dir(pkg.GoFiles[0])
	p.Name = pkg.Name
	p.pkg = pkg

	fieldVars := make(map[string]map[string]*types.Var)
	for ident, obj := range pkg.TypesInfo.Defs {
		// skip unexported structs, and exported fields by looking for a nil
		// parent scope.
		if obj == nil || !obj.Exported() || obj.Parent() == nil {
			continue
		}

		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		strct, ok := named.Underlying().(*types.Struct)
		if !ok {
			continue
		}

		fields := make(map[string]*types.Var)
		for i := 0; i < strct.NumFields(); i++ {
			f := strct.Field(i)
			if f.Exported() {
				fields[f.Name()] = f
			}
		}
		fieldVars[ident.Name] = fields
	}

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl := declAsTypeGenDecl(decl)
			if genDecl == nil {
				continue
			}

			for _, spec := range genDecl.Specs {
				spec := specAsExpectedTypeSpec(spec)
				if spec == nil {
					continue
				}

				// godoc may be on the GenDecl or the TypeSpec
				doc := spec.Doc
				if doc == nil {
					doc = genDecl.Doc
				}

				structType, ok := spec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if !containsMogStructAnnotation(doc) {
					continue
				}

				typedFields := make([]typedField, len(structType.Fields.List))
				for i, f := range structType.Fields.List {
					name, err := fieldName(f)
					if err != nil {
						return p, err
					}

					typedFields[i] = typedField{
						Field: f,
						Var:   fieldVars[spec.Name.Name][name],
					}

				}

				p.Structs[spec.Name.Name] = structDecl{
					Doc:    doc.List,
					Fields: typedFields,
				}
			}
		}
	}

	return p, nil
}

// TODO: trim this if All isn't needed
var modeLoadAll = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes

func packageLoadErrors(pkg *packages.Package) error {
	if len(pkg.Errors) == 0 {
		return nil
	}

	buf := new(strings.Builder)
	for _, err := range pkg.Errors {
		buf.WriteString("\n")
		buf.WriteString(err.Error())
	}
	buf.WriteString("\n")
	return fmt.Errorf("package %s has errors: %s", pkg.PkgPath, buf.String())
}

func declAsTypeGenDecl(o ast.Decl) *ast.GenDecl {
	if o == nil {
		return nil
	}
	decl, ok := o.(*ast.GenDecl)
	if !ok {
		return nil
	}
	if decl.Tok != token.TYPE {
		return nil
	}
	return decl
}

func specAsExpectedTypeSpec(s ast.Spec) *ast.TypeSpec {
	spec, ok := s.(*ast.TypeSpec)
	if !ok {
		return nil
	}
	if !spec.Name.IsExported() {
		return nil
	}
	return spec
}

// containsMogStructAnnotation scans the lines in the doc comment group and returns
// true if one of the lines contains the comment which identifies the struct as one
// that should be used for the source of type conversion.
func containsMogStructAnnotation(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	return structAnnotationIndex(doc.List) != -1
}

func structAnnotationIndex(doc []*ast.Comment) int {
	for i, line := range doc {
		text := strings.TrimSpace(strings.TrimLeft(line.Text, "/"))
		if text == "mog annotation:" {
			return i
		}
	}
	return -1
}

type targetPkg struct {
	Structs map[string]targetStruct
}

type targetStruct struct {
	Name   string
	Fields []*types.Var
}

func loadTargetStructs(names []string, tags string) (map[string]targetPkg, error) {
	mode := packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName
	cfg := &packages.Config{
		Mode: mode,
	}
	if tags == "" {
		tags = os.Getenv("GOTAGS")
	}
	if tags != "" {
		cfg.BuildFlags = []string{fmt.Sprintf("-tags=%s", tags)}
	}

	pkgs, err := packages.Load(cfg, names...)
	if err != nil {
		return nil, err
	}

	result := make(map[string]targetPkg, len(names))
	for _, pkg := range pkgs {
		if err := packageLoadErrors(pkg); err != nil {
			return nil, err
		}

		structs := map[string]targetStruct{}
		for ident, obj := range pkg.TypesInfo.Defs {
			// skip unexported structs, and exported fields by looking for a nil
			// parent scope.
			if obj == nil || !obj.Exported() || obj.Parent() == nil {
				continue
			}

			named, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}

			strct, ok := named.Underlying().(*types.Struct)
			if !ok {
				continue
			}

			var fields []*types.Var
			for i := 0; i < strct.NumFields(); i++ {
				f := strct.Field(i)
				if f.Exported() {
					fields = append(fields, f)
				}
			}
			structs[ident.Name] = targetStruct{Name: ident.Name, Fields: fields}
		}
		result[pkg.PkgPath] = targetPkg{Structs: structs}
	}
	return result, nil
}
