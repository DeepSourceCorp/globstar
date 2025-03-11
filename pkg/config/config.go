package config

import (
	"fmt"
	"os"

	"github.com/gobwas/glob"
	"gopkg.in/yaml.v3"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityError    Severity = "error"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityError, SeverityWarning, SeverityInfo:
		return true
	}
	return false
}

type Category string

const (
	CategoryStyle       Category = "style"
	CategoryBugRisk     Category = "bug-risk"
	CategoryAntipattern Category = "antipattern"
	CategoryPerformance Category = "performance"
	CategorySecurity    Category = "security"
)

func (c Category) IsValid() bool {
	switch c {
	case CategoryStyle, CategoryBugRisk, CategoryAntipattern, CategoryPerformance, CategorySecurity:
		return true
	}
	return false
}

type FailureConfig struct {
	ExitCode   int                 `yaml:"exitCode"`
	SeverityIn []Severity          `yaml:"severityIn"`
	CategoryIn []Category          `yaml:"categoryIn"`
	MetadataIn []map[string]string `yaml:"metadataIn"`
}

func (fc *FailureConfig) PopulateDefaults() {
	if fc.ExitCode == 0 {
		fc.ExitCode = 1
	}

	if len(fc.SeverityIn) == 0 {
		fc.SeverityIn = []Severity{SeverityCritical}
	}

	if len(fc.CategoryIn) == 0 {
		fc.CategoryIn = []Category{CategoryBugRisk}
	}
}

type Config struct {
	CheckerDir       string        `yaml:"checkerDir"`
	EnabledCheckers  []string      `yaml:"enabledCheckers"`
	DisabledCheckers []string      `yaml:"disabledCheckers"`
	TargetDirs       []string      `yaml:"targetDirs"`
	ExcludePatterns  []string      `yaml:"excludePatterns"`
	FailWhen         FailureConfig `yaml:"failWhen"`

	excludedGlobs []glob.Glob
}

func NewConfigFromFile(path string) (*Config, error) {
	c := &Config{}
	c.PopulateDefaults()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c, nil // ignore if file does not exist
	}

	s, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(s, c)
	if err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (config *Config) PopulateDefaults() {
	if config.CheckerDir == "" {
		config.CheckerDir = ".globstar"
	}

	config.FailWhen.PopulateDefaults()
}

func (config *Config) Validate() error {
	if err := config.validateExcludePatterns(); err != nil {
		return err
	}
	if err := config.validateFailureConfig(); err != nil {
		return err
	}
	return nil
}

func (config *Config) validateExcludePatterns() error {
	for _, pattern := range config.ExcludePatterns {
		p, err := glob.Compile(pattern)
		if err != nil {
			return fmt.Errorf("Could not validate pattern %s.", pattern)
		}

		config.excludedGlobs = append(config.excludedGlobs, p)
	}
	return nil
}

func (config *Config) validateFailureConfig() error {
	if config.FailWhen.ExitCode < 0 {
		return fmt.Errorf("exitCode must be a non-negative integer")
	}

	for _, severity := range config.FailWhen.SeverityIn {
		if !severity.IsValid() {
			return fmt.Errorf("invalid severity: %s", severity)
		}
	}

	for _, category := range config.FailWhen.CategoryIn {
		if !category.IsValid() {
			return fmt.Errorf("invalid category: %s", category)
		}
	}

	return nil
}

func (config *Config) ShouldExcludePath(path string) bool {
	for _, g := range config.excludedGlobs {
		if g.Match(path) {
			return true
		}
	}
	return false
}

func (config *Config) AddExcludePatterns(patterns ...string) error {
	config.ExcludePatterns = append(config.ExcludePatterns, patterns...)

	for _, pattern := range patterns {
		p, err := glob.Compile(pattern)
		if err != nil {
			return fmt.Errorf("Could not validate pattern %s.", pattern)
		}

		config.excludedGlobs = append(config.excludedGlobs, p)
	}

	return nil
}
