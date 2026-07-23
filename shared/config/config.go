// Package config loads YAML configuration files and resolves ${ENV_VAR}
// placeholders against the process environment, so secrets never live in
// the repository (see spec sections "Конфигурация" and "Безопасность").
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	apperrors "streampass/shared/errors"
)

var placeholderPattern = regexp.MustCompile(`\$\{([A-Z0-9_]+)\}`)

// FileLoader loads a YAML config file from disk, resolving ${ENV_VAR}
// placeholders found in scalar string values.
type FileLoader struct {
	path string
}

// NewFileLoader builds a loader for the given YAML file path.
func NewFileLoader(path string) *FileLoader {
	return &FileLoader{path: path}
}

// Config is a resolved, read-only configuration tree with typed getters.
// A thin wrapper rather than reflection-based struct binding — simplest
// option that still gives callers safe, explicit access (KISS).
type Config struct {
	root yamlNode
}

// Load reads the file, resolves placeholders and returns a Config.
func (l *FileLoader) Load() (*Config, error) {
	raw, err := os.ReadFile(l.path)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeNotFound, "config file not found: "+l.path, err)
	}
	resolved := resolvePlaceholders(string(raw))
	root := parseYAML(resolved)
	return &Config{root: root}, nil
}

func resolvePlaceholders(content string) string {
	return placeholderPattern.ReplaceAllStringFunc(content, func(match string) string {
		name := placeholderPattern.FindStringSubmatch(match)[1]
		if v, ok := os.LookupEnv(name); ok {
			return v
		}
		return "" // absent env var resolves to empty string, never the literal placeholder text —
		// a literal "${VAR}" being used as real data (e.g. as a password) is worse than empty
	})
}

// String returns a string value at a dotted path (e.g. "database.host").
func (c *Config) String(path string) (string, error) {
	v, ok := c.root.get(path)
	if !ok {
		return "", apperrors.New(apperrors.CodeNotFound, "config key not found: "+path)
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v), nil
	}
	return s, nil
}

// StringOr returns a string value or a fallback default.
func (c *Config) StringOr(path, def string) string {
	v, err := c.String(path)
	if err != nil || v == "" {
		return def
	}
	return v
}

// Int returns an int value at a dotted path.
func (c *Config) Int(path string) (int, error) {
	v, ok := c.root.get(path)
	if !ok {
		return 0, apperrors.New(apperrors.CodeNotFound, "config key not found: "+path)
	}
	switch t := v.(type) {
	case int:
		return t, nil
	case string:
		i, err := strconv.Atoi(t)
		if err != nil {
			return 0, apperrors.Wrap(apperrors.CodeInvalidInput, "config key is not an int: "+path, err)
		}
		return i, nil
	default:
		return 0, apperrors.New(apperrors.CodeInvalidInput, "config key is not an int: "+path)
	}
}

// IntOr returns an int value or a fallback default.
func (c *Config) IntOr(path string, def int) int {
	v, err := c.Int(path)
	if err != nil {
		return def
	}
	return v
}

// Duration returns a Go duration parsed from a string value (e.g. "15m").
func (c *Config) Duration(path string) (time.Duration, error) {
	s, err := c.String(path)
	if err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInvalidInput, "config key is not a duration: "+path, err)
	}
	return d, nil
}

// DurationOr returns a duration value or a fallback default.
func (c *Config) DurationOr(path string, def time.Duration) time.Duration {
	d, err := c.Duration(path)
	if err != nil {
		return def
	}
	return d
}

// Bool returns a bool value at a dotted path.
func (c *Config) Bool(path string) (bool, error) {
	v, ok := c.root.get(path)
	if !ok {
		return false, apperrors.New(apperrors.CodeNotFound, "config key not found: "+path)
	}
	b, ok := v.(bool)
	if !ok {
		return false, apperrors.New(apperrors.CodeInvalidInput, "config key is not a bool: "+path)
	}
	return b, nil
}

// BoolOr returns a bool value or a fallback default.
func (c *Config) BoolOr(path string, def bool) bool {
	b, err := c.Bool(path)
	if err != nil {
		return def
	}
	return b
}
