package parser

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// StructInfo holds information about the parsed struct
type StructInfo struct {
	PackageName string
	Name        string
	Dir         string // Directory containing the source file
	Imports     []string
	Fields      []FieldInfo
}

// FieldInfo holds information about a single field
type FieldInfo struct {
	Name string
	Type string
	Tags string
}

// Parse loads the package and finds the specified struct
func Parse(pattern string, structName string) (*StructInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contained errors")
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok.String() != "type" {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Name.Name != structName {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						return nil, fmt.Errorf("%s is not a struct", structName)
					}

					// Get directory of the file containing the struct
					filePos := pkg.Fset.Position(typeSpec.Pos())
					dir := filepath.Dir(filePos.Filename)

					info := &StructInfo{
						PackageName: pkg.Name,
						Name:        structName,
						Dir:         dir,
						Fields:      make([]FieldInfo, 0),
						Imports:     make([]string, 0),
					}

					importMap := make(map[string]bool)

					for _, field := range structType.Fields.List {
						// Extract imports from type
						if field.Type != nil {
							typ := pkg.TypesInfo.TypeOf(field.Type)
							collectImports(typ, importMap)
						}

						// Handle embedded fields or multiple fields per line (e.g. x, y int)
						if len(field.Names) == 0 {
							// Embedded field
							// For now, we might skip or handle simple cases.
							// Let's get the type name as the field name for embedded fields.
							typeName := types.ExprString(field.Type)
							// Remove package prefix if present for simple name
							parts := strings.Split(typeName, ".")
							fieldName := parts[len(parts)-1]

							info.Fields = append(info.Fields, FieldInfo{
								Name: fieldName,
								Type: typeName,
								Tags: getTag(field.Tag),
							})
							continue
						}

						for _, name := range field.Names {
							info.Fields = append(info.Fields, FieldInfo{
								Name: name.Name,
								Type: types.ExprString(field.Type),
								Tags: getTag(field.Tag),
							})
						}
					}

					for imp := range importMap {
						info.Imports = append(info.Imports, imp)
					}

					return info, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("struct %s not found in pattern %s", structName, pattern)
}

func collectImports(typ types.Type, imports map[string]bool) {
	switch t := typ.(type) {
	case *types.Named:
		if t.Obj() != nil && t.Obj().Pkg() != nil {
			imports[t.Obj().Pkg().Path()] = true
		}
		// Also check type arguments if generics are used (future proofing)
		if t.TypeArgs() != nil {
			for i := 0; i < t.TypeArgs().Len(); i++ {
				collectImports(t.TypeArgs().At(i), imports)
			}
		}
	case *types.Pointer:
		collectImports(t.Elem(), imports)
	case *types.Slice:
		collectImports(t.Elem(), imports)
	case *types.Array:
		collectImports(t.Elem(), imports)
	case *types.Map:
		collectImports(t.Key(), imports)
		collectImports(t.Elem(), imports)
	case *types.Chan:
		collectImports(t.Elem(), imports)
	}
}

func getTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	return tag.Value
}
