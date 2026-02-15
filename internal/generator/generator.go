package generator

import (
	"bytes"
	"dtogen/internal/parser"
	"dtogen/internal/types"
	"embed"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

//go:embed templates/dto.tmpl
var defaultTemplateFS embed.FS

type Config struct {
	PackageName    string
	StructName     string
	DTOName        string
	Excludes       map[string]bool
	Includes       map[string]bool
	Renames        map[string]string
	Filters        []types.Filter
	Imports        []string
	AddFields      []string
	ExcludeImports []string
	TemplatePath   string
}

type templateData struct {
	PackageName string
	StructName  string
	DTOName     string
	Imports     []string
	Fields      []FieldData
	AddFields   []string
	Filters     []types.Filter
}

type FieldData struct {
	Name       string // DTO field name
	SourceName string // Original struct field name
	Type       string
	Tags       string
}

func Generate(info *parser.StructInfo, cfg Config) ([]byte, error) {
	// Filter and map fields
	var fields []FieldData
	for _, f := range info.Fields {
		// If Includes is provided, only include fields in the list
		if len(cfg.Includes) > 0 {
			if !cfg.Includes[f.Name] {
				continue
			}
		} else if cfg.Excludes[f.Name] {
			// Otherwise check excludes
			continue
		}

		dtoFieldName := f.Name
		if newName, ok := cfg.Renames[f.Name]; ok {
			dtoFieldName = newName
		}

		fields = append(fields, FieldData{
			Name:       dtoFieldName,
			SourceName: f.Name,
			Type:       f.Type,
			Tags:       f.Tags,
		})
	}

	excluded_imports := make(map[string]struct{})
	for _, item := range cfg.ExcludeImports {
		excluded_imports[item] = struct{}{}
	}

	imports := cfg.Imports[0:]
	for _, info_import := range info.Imports {
		if _, found := excluded_imports[info_import]; !found {
			imports = append(imports, info_import)
		}
	}

	// imports := append(cfg.Imports, info.Imports...)

	data := templateData{
		PackageName: cfg.PackageName,
		StructName:  cfg.StructName,
		DTOName:     cfg.DTOName,
		Imports:     imports,
		Fields:      fields,
		AddFields:   cfg.AddFields,
		Filters:     cfg.Filters,
	}

	var tmpl *template.Template
	var err error

	if cfg.TemplatePath != "" {
		content, err := os.ReadFile(cfg.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file: %w", err)
		}
		tmpl, err = template.New("custom").Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse custom template: %w", err)
		}
	} else {
		// Use default template
		// Note: The path in ParseFS must match the embed path
		tmpl, err = template.ParseFS(defaultTemplateFS, "templates/dto.tmpl")
		if err != nil {
			return nil, fmt.Errorf("failed to parse default template: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code for debugging if formatting fails
		return buf.Bytes(), fmt.Errorf("failed to format generated code: %w", err)
	}

	return formatted, nil
}
