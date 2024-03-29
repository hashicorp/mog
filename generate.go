// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/rboyer/safeio"
)

func generateFiles(cfg config, targets map[string]targetPkg) error {
	byOutput := configsByOutput(cfg.Structs)

	for _, group := range byOutput {
		var decls []ast.Decl
		imports := newImports()

		for _, sourceStruct := range group {
			t := targets[sourceStruct.Target.Package].Structs[sourceStruct.Target.Struct]
			if t.Name == "" {
				return fmt.Errorf("failed to locate target %v for %v", sourceStruct.Target, sourceStruct.Source)
			}

			gen, err := generateConversion(sourceStruct, t, imports)
			if err != nil {
				return fmt.Errorf("failed to generate conversion for %v: %w", sourceStruct.Source, err)
			}
			decls = append(decls, gen.To, gen.From)

			// TODO: generate round trip testcase
		}

		fset := &token.FileSet{}
		file := &ast.File{Name: &ast.Ident{Name: cfg.SourcePkg.Name}}
		output := filepath.Join(cfg.SourcePkg.Path, group[0].Output)

		// Add all imports as the first declaration
		// TODO: dedupe imports, handle conflicts
		file.Decls = append([]ast.Decl{imports.Decl()}, decls...)

		if err := astWriteToFile(output, fset, file); err != nil {
			return fmt.Errorf("failed to write generated code to %v: %w", output, err)
		}
	}
	return nil
}

// configsByOutput sorts and groups the configs by the Output filename. Each
// group is sorted by name of struct.
func configsByOutput(cfgs []structConfig) [][]structConfig {
	if len(cfgs) == 0 {
		return nil
	}

	var result [][]structConfig
	sort.Slice(cfgs, func(i, j int) bool {
		if cfgs[i].Output == cfgs[j].Output {
			return cfgs[i].Source < cfgs[j].Source
		}
		return cfgs[i].Output < cfgs[j].Output
	})

	var group []structConfig
	output := cfgs[0].Output
	for _, c := range cfgs {
		if c.Output != output {
			result = append(result, group)
			group = []structConfig{}
			output = c.Output
		}

		group = append(group, c)
	}
	return append(result, group)
}

var (
	varNameSource          = "s"
	varNameTarget          = "t"
	varNamePlaceholder     = "x"
	varNameElemPlaceholder = "y"
)

func generateConversion(cfg structConfig, t targetStruct, imports *imports) (generated, error) {
	var g generated

	imports.Add("", cfg.Target.Package)

	targetType := &ast.SelectorExpr{
		X:   &ast.Ident{Name: path.Base(imports.AliasFor(cfg.Target.Package))},
		Sel: &ast.Ident{Name: cfg.Target.Struct},
	}

	to := generateToFunc(cfg, targetType)
	from := generateFromFunc(cfg, targetType)

	var errs []error

	// TODO: would it make sense to store the fields as a map instead of building it here?
	sourceFields := sourceFieldMap(cfg.Fields)
	for _, field := range t.Fields {
		name := field.Name()

		// TODO: test case to include ignored field
		if _, contains := cfg.IgnoreFields[name]; contains {
			continue
		}

		sourceField, ok := sourceFields[name]
		if !ok {
			msg := "struct %v is missing field %v. Add the missing field or exclude it using ignore-fields."
			errs = append(errs, fmt.Errorf(msg, cfg.Source, name))
			continue
		}

		srcExpr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: varNameSource},
			Sel: &ast.Ident{Name: sourceField.SourceName},
		}
		targetExpr := &ast.SelectorExpr{
			X:   &ast.Ident{Name: varNameTarget},
			Sel: &ast.Ident{Name: name},
		}

		if sourceField.FuncTo != "" || sourceField.FuncFrom != "" {
			to.Body.List = append(to.Body.List, newAssignStmtUserFunc(
				targetExpr,
				srcExpr,
				sourceField.FuncTo,
			))
			from.Body.List = append(from.Body.List, newAssignStmtUserFunc(
				srcExpr,
				targetExpr,
				sourceField.FuncFrom,
			))
			continue
		}

		targetTypeExpr := typeToExpr(field.Type(), imports, false)
		if targetTypeExpr == nil {
			msg := "struct %v field %v is not a supported target type yet: %T"
			errs = append(errs, fmt.Errorf(msg, cfg.Source, name, field.Type()))
			continue
		}

		assignErrFn := func(err error) {
			if err == nil {
				errs = append(errs, fmt.Errorf(
					"struct %v field %v is not convertible to target",
					cfg.Source, name,
				))
			} else {
				errs = append(errs, fmt.Errorf(
					"struct %v field %v is not convertible to target: %w",
					cfg.Source, name, err,
				))
			}
		}

		// the assignmentKind is <target> := <source> so target==LHS source==RHS
		rawKind, ok := computeAssignment(field.Type(), sourceField.SourceType)
		if !ok {
			assignErrFn(nil)
			continue
		}

		switch kind := rawKind.(type) {
		case *singleAssignmentKind:
			to.Body.List = append(to.Body.List, newAssignStmt(
				targetExpr,
				targetTypeExpr,
				srcExpr,
				sourceField.SourceExpr,
				sourceField.ConvertFuncName(DirTo),
				kind.Direct,
				kind.Convert,
			))
			from.Body.List = append(from.Body.List, newAssignStmt(
				srcExpr,
				sourceField.SourceExpr,
				targetExpr,
				targetTypeExpr,
				sourceField.ConvertFuncName(DirFrom),
				kind.Direct,
				kind.Convert,
			))
		case *sliceAssignmentKind:
			targetElemTypeElem := typeToExpr(kind.LeftElem, imports, true)
			if targetElemTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported slice element type %T", kind.LeftElem))
				continue
			}

			sourceElemTypeElem := typeToExpr(kind.RightElem, imports, true)
			if sourceElemTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported slice element type %T", kind.RightElem))
				continue
			}

			to.Body.List = append(to.Body.List, newAssignStmtSlice(
				targetExpr,
				targetTypeExpr,
				targetElemTypeElem,
				srcExpr,
				sourceElemTypeElem,
				sourceField.ConvertFuncName(DirTo),
				kind.ElemDirect,
				kind.ElemConvert,
			))
			from.Body.List = append(from.Body.List, newAssignStmtSlice(
				srcExpr,
				sourceField.SourceExpr,
				sourceElemTypeElem,
				targetExpr,
				targetElemTypeElem,
				sourceField.ConvertFuncName(DirFrom),
				kind.ElemDirect,
				kind.ElemConvert,
			))
		case *mapAssignmentKind:
			targetKeyTypeElem := typeToExpr(kind.LeftKey, imports, true)
			if targetKeyTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported map key type %T", kind.LeftKey))
				continue
			}

			targetElemTypeElem := typeToExpr(kind.LeftElem, imports, true)
			if targetElemTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported map value type %T", kind.LeftElem))
				continue
			}

			sourceKeyTypeElem := typeToExpr(kind.RightKey, imports, true)
			if sourceKeyTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported map key type %T", kind.RightKey))
				continue
			}

			sourceElemTypeElem := typeToExpr(kind.RightElem, imports, true)
			if sourceElemTypeElem == nil {
				assignErrFn(fmt.Errorf("unsupported map value type %T", kind.RightElem))
				continue
			}

			to.Body.List = append(to.Body.List, newAssignStmtMap(
				targetExpr,
				targetTypeExpr,
				targetElemTypeElem,
				srcExpr,
				sourceElemTypeElem,
				sourceField.ConvertFuncName(DirTo),
				kind.ElemDirect,
				kind.ElemConvert,
			))
			from.Body.List = append(from.Body.List, newAssignStmtMap(
				srcExpr,
				sourceField.SourceExpr,
				sourceElemTypeElem,
				targetExpr,
				targetElemTypeElem,
				sourceField.ConvertFuncName(DirFrom),
				kind.ElemDirect,
				kind.ElemConvert,
			))
		}
	}

	g.To = to
	g.From = from

	return g, fmtErrors("failed to generate", errs)
}

func sourceFieldMap(fields []fieldConfig) map[string]fieldConfig {
	result := make(map[string]fieldConfig, len(fields))
	for _, field := range fields {
		key := field.SourceName
		if field.TargetName != "" {
			key = field.TargetName
		}
		result[key] = field
	}
	return result
}

type generated struct {
	To   *ast.FuncDecl
	From *ast.FuncDecl

	// TODO: RoundTripTest *ast.FuncDecl
}

func generateToFunc(cfg structConfig, targetType *ast.SelectorExpr) *ast.FuncDecl {
	funcName := cfg.ConvertFuncName(DirTo)

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: funcName},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: varNameSource}},
						Type:  newPointerTo(cfg.Source),
					},
					{
						Names: []*ast.Ident{{Name: varNameTarget}},
						Type: &ast.StarExpr{
							X: targetType,
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			newIfNilReturn(varNameSource),
			// TODO: fill in contents here
		}},
	}
}

func generateFromFunc(cfg structConfig, targetType *ast.SelectorExpr) *ast.FuncDecl {
	funcName := cfg.ConvertFuncName(DirFrom)

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: funcName},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: varNameTarget}},
						Type: &ast.StarExpr{
							X: targetType,
						},
					},
					{
						Names: []*ast.Ident{{Name: varNameSource}},
						Type:  newPointerTo(cfg.Source),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				newIfNilReturn(varNameSource),
				// TODO: fill in contents here
			},
		},
	}
}

func astWriteToFile(path string, fset *token.FileSet, file *ast.File) error {
	out, err := astToBytes(fset, file)
	if err != nil {
		return err
	}

	return writeFile(path, out)
}

func astToBytes(fset *token.FileSet, file *ast.File) ([]byte, error) {
	// Pretty print the AST node first.
	printConfig := &printer.Config{Mode: printer.TabIndent}
	var buf bytes.Buffer
	if err := printConfig.Fprint(&buf, fset, file); err != nil {
		return nil, err
	}
	out := buf.Bytes()

	// Now take a trip through "gofmt"
	formatted, err := format.Source(out)
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

// TODO: write build tags
func writeFile(output string, contents []byte) error {
	fh, err := safeio.OpenFile(output, 0666)
	if err != nil {
		return err
	}
	defer fh.Close()

	if _, err := fmt.Fprint(fh, "// Code generated by mog. DO NOT EDIT.\n\n"); err != nil {
		return err
	}
	if _, err := fh.Write(contents); err != nil {
		return err
	}

	return fh.Commit()
}

type imports struct {
	byPkgPath map[string]string   // package => alias(or default)
	byAlias   map[string]string   // alias(or default) => package
	hasAlias  map[string]struct{} // package is using a non-default name
}

func newImports() *imports {
	return &imports{
		byPkgPath: make(map[string]string),
		byAlias:   make(map[string]string),
		hasAlias:  make(map[string]struct{}),
	}
}

// Add an import with an optional alias. If no alias is specified, the default
// alias will be path.Base(). The alias for a package should always be looked up
// from AliasFor.
//
// TODO: remove alias arg?
func (i *imports) Add(alias string, pkgPath string) {
	if _, exists := i.byPkgPath[pkgPath]; exists {
		return
	}

	hasAlias := false
	if alias == "" {
		alias = path.Base(pkgPath)
	} else {
		hasAlias = true
	}

	_, exists := i.byAlias[alias]
	for n := 2; exists; n++ {
		alias = alias + strconv.Itoa(n)
		_, exists = i.byAlias[alias]
	}

	i.byPkgPath[pkgPath] = alias
	i.byAlias[alias] = pkgPath
	if hasAlias {
		i.hasAlias[pkgPath] = struct{}{}
	} else {
		delete(i.hasAlias, pkgPath)
	}
}

func (i *imports) AliasFor(pkgPath string) string {
	return i.byPkgPath[pkgPath]
}

func (i *imports) Decl() *ast.GenDecl {
	decl := &ast.GenDecl{Tok: token.IMPORT}

	paths := make([]string, 0, len(i.byPkgPath))
	for pkgPath := range i.byPkgPath {
		paths = append(paths, pkgPath)
	}
	sort.Strings(paths)

	for _, pkgPath := range paths {
		imprt := &ast.ImportSpec{
			Path: &ast.BasicLit{Value: strconv.Quote(pkgPath), Kind: token.STRING},
		}

		if _, ok := i.hasAlias[pkgPath]; ok {
			imprt.Name = &ast.Ident{Name: i.byPkgPath[pkgPath]}
		}

		decl.Specs = append(decl.Specs, imprt)
	}
	return decl
}
