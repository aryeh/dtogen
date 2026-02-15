package config

import (
	"dtogen/internal/types"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Global GlobalConfig `yaml:"global"`
	DTOs   []DTOConfig  `yaml:"dtos"`
}

type GlobalConfig struct {
	OutputDir      string   `yaml:"output_dir"`
	Imports        []string `yaml:"imports"`
	ExcludeImports []string `yaml:"exclude_imports"`
	Source         string   `yaml:"source"`
}

type DTOConfig struct {
	Type      string            `yaml:"type"`
	Source    string            `yaml:"source"`
	Name      string            `yaml:"name"`
	Output    string            `yaml:"output"`
	Excludes  []string          `yaml:"excludes"`
	Includes  []string          `yaml:"includes"`
	AddFields []string          `yaml:"add_fields"`
	Renames   map[string]string `yaml:"renames"`
	Filters   []types.Filter    `yaml:"filters"`
	Template  string            `yaml:"template"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func GenerateSample(path string) error {
	sample := `
# Global settings applied to all DTOs unless overridden
global:
  output_dir: "./generated"
  imports:
	  - "godev/types"

dtos:
  - type: User
    source: "./models"
    name: UserDTO
    output: user_dto.go
    excludes:
      - Password
    # includes: [ID, Username]
    # renames:
    #   Username: Login
    # filters:
    #   - when: "IsConfidential"
    #     do: |
    #       Cmd 1
    #       Cmd 2

`
	return os.WriteFile(path, []byte(sample), 0644)
}
