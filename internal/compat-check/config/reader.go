package config

import (
	"fmt"
	"os"
)

// ConfigReader is the sole I/O object that touches contract-tests.yaml on disk. It is an
// empty-pipe: read bytes, no transformation — proven by component scenario CS3, not unit-tested
// (module-tree.md node → file map).
type ConfigReader struct{}

// NewConfigReader constructs the (dependency-free) reader.
func NewConfigReader() ConfigReader {
	return ConfigReader{}
}

// Read returns the raw bytes of the config file at path. Any failure (missing file, unreadable,
// permission denied) collapses to ErrConfigInvalid — error.code CONFIG_INVALID, exit 2
// (contracts.md "Error model").
func (r ConfigReader) Read(path string) (RawConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %s", ErrConfigInvalid, path, err)
	}
	return RawConfig(data), nil
}
