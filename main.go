package main

import (
	"dtogen/internal/config"
	"dtogen/internal/generator"
	"dtogen/internal/parser"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/pflag"
)

func main() {
	var (
		src        string
		typeName   string
		dtoName    string
		out        string
		outDir     string
		excludes   string
		includes   string
		renames    []string
		tmplPath   string
		configFile string
		initSample bool
	)

	pflag.StringVarP(&src, "src", "s", ".", "Source file or package pattern")
	pflag.StringVarP(&typeName, "type", "t", "", "Struct name to generate from (Required for CLI mode)")
	pflag.StringVarP(&dtoName, "name", "n", "", "Output DTO struct name (Default: [Type]DTO)")
	pflag.StringVarP(&out, "out", "o", "", "Output file name (Default: [type]_dto.go)")
	pflag.StringVarP(&outDir, "dir", "d", "", "Output directory (Default: .)")
	pflag.StringVarP(&excludes, "exclude", "e", "", "Comma-separated fields to exclude")
	pflag.StringVarP(&includes, "include", "i", "", "Comma-separated fields to include (overrides exclude)")
	pflag.StringSliceVarP(&renames, "replace", "r", nil, "Field renames in format Source:Target (can be repeated)")
	pflag.StringVar(&tmplPath, "template", "", "Path to custom template file")
	pflag.StringVar(&configFile, "config", "", "Path to configuration file")
	pflag.BoolVar(&initSample, "init", false, "Generate sample configuration file (dtogen_sample.yaml)")

	pflag.Parse()

	if initSample {
		if err := config.GenerateSample("dtogen_sample.yaml"); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating sample config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Generated dtogen_sample.yaml")
		return
	}

	// Check for config file
	if configFile != "" {
		runConfig(configFile)
		return
	}

	// CLI Mode
	if typeName == "" {
		pflag.Usage()
		os.Exit(1)
	}

	if dtoName == "" {
		dtoName = typeName + "DTO"
	}

	if out == "" {
		out = strings.ToLower(typeName) + "_dto.go"
	}

	// Prepare config
	excludeMap := make(map[string]bool)
	if excludes != "" {
		for _, f := range strings.Split(excludes, ",") {
			excludeMap[strings.TrimSpace(f)] = true
		}
	}

	includeMap := make(map[string]bool)
	if includes != "" {
		for _, f := range strings.Split(includes, ",") {
			includeMap[strings.TrimSpace(f)] = true
		}
	}

	renameMap := make(map[string]string)
	for _, r := range renames {
		parts := strings.Split(r, ":")
		if len(parts) == 2 {
			renameMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	cfg := generator.Config{
		StructName:   typeName,
		DTOName:      dtoName,
		Excludes:     excludeMap,
		Includes:     includeMap,
		Renames:      renameMap,
		TemplatePath: tmplPath,
	}

	if err := runGenerator(src, outDir, out, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runConfig(path string) {
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	for _, dto := range cfg.DTOs {
		// Merge global settings
		outDir := cfg.Global.OutputDir

		// Defaults
		source := dto.Source
		if source == "" {
			source = cfg.Global.Source
			if source == "" {
				source = "."
			}
		}

		name := dto.Name
		if name == "" {
			name = dto.Type + "DTO"
		}

		out := dto.Output
		if out == "" {
			// out = strings.ToLower(dto.Type) + "_dto.go"
			out = pascalCaseToSnakeCase(name) + ".go"
		}

		excludeMap := make(map[string]bool)
		for _, f := range dto.Excludes {
			excludeMap[f] = true
		}

		includeMap := make(map[string]bool)
		for _, f := range dto.Includes {
			includeMap[f] = true
		}

		genCfg := generator.Config{
			StructName:     dto.Type,
			DTOName:        name,
			Excludes:       excludeMap,
			Includes:       includeMap,
			Renames:        dto.Renames,
			Filters:        dto.Filters,
			AddFields:      dto.AddFields,
			TemplatePath:   dto.Template,
			Imports:        cfg.Global.Imports,
			ExcludeImports: cfg.Global.ExcludeImports,
		}

		if err := runGenerator(source, outDir, out, genCfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", dto.Type, err)
			// Continue with other DTOs? Or exit? Let's continue but report error.
		}
	}
}

// pascalCaseToSnakeCase converts a PascalCase string to snake_case.
func pascalCaseToSnakeCase(s string) string {
	var builder strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add an underscore if it's not the first character
			// and the previous character was lowercase or the next is also uppercase (for acronyms)
			if i > 0 && (unicode.IsLower(rune(s[i-1])) /* || (i+1 < len(s) && unicode.IsUpper(rune(s[i+1]))) */) {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(r))
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func runGenerator(src, outDir, out string, cfg generator.Config) error {
	// Construct full output path
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		out = filepath.Join(outDir, out)

		// remove the dest file.
		if err := os.Remove(out); err != nil {
			fmt.Printf("Failed removing file before starting [%s]: %v\n", out, err)
		}

	}

	// Parse source
	info, err := parser.Parse(src, cfg.StructName)
	if err != nil {
		return fmt.Errorf("parsing source: %w", err)
	}

	cfg.PackageName = info.PackageName

	// Generate code
	code, err := generator.Generate(info, cfg)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	if outDir == "" {
		// Default to source directory
		out = filepath.Join(info.Dir, out)
	}

	// Write to file
	if err := os.WriteFile(out, code, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	// fmt.Printf("Generated %s\n", out)
	return nil
}
