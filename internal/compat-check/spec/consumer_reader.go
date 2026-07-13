package spec

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// ErrSpecUnreadable is raised when the consumer OpenAPI file at spec_path is missing or
// unreadable — error.code SPEC_UNREADABLE, exit 2 (contracts.md "Error model").
var ErrSpecUnreadable = errors.New("consumer spec unreadable")

// ErrSpecParse is raised when the consumer document reads but does not parse/validate as
// OpenAPI 3.x — error.code SPEC_PARSE_ERROR, exit 2 (contracts.md "Error model").
var ErrSpecParse = errors.New("consumer spec parse error")

// ConsumerSpecReader is the sole I/O object that loads the consumer OpenAPI document from disk.
// Pipe — read the file + honest-reuse kin-openapi to parse and resolve local/external $refs; no
// transformation of its own. Not unit-tested (module-tree.md node -> file map); proven by
// component scenarios CS4 (unreadable) / CS5 (unparseable).
type ConsumerSpecReader struct{}

// NewConsumerSpecReader constructs the (dependency-free) reader.
func NewConsumerSpecReader() ConsumerSpecReader {
	return ConsumerSpecReader{}
}

// Load reads the consumer file at path and honest-reuses kin-openapi to parse it, resolving
// local and external $refs (IsExternalRefsAllowed). Missing/unreadable file -> ErrSpecUnreadable.
// Content that does not parse or does not validate as OpenAPI 3.x -> ErrSpecParse.
func (r ConsumerSpecReader) Load(path string) (Spec, error) {
	// A dedicated readability check first — distinguishes "can't read the file"
	// (ErrSpecUnreadable) from "read fine, but the content is bad OpenAPI" (ErrSpecParse); the
	// kin-openapi loader below re-reads the file to get its own $ref-relative-path handling.
	if _, err := os.ReadFile(path); err != nil {
		return Spec{}, fmt.Errorf("%w: %s: %s", ErrSpecUnreadable, path, err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("%w: %s: %s", ErrSpecParse, path, err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		return Spec{}, fmt.Errorf("%w: %s: %s", ErrSpecParse, path, err)
	}

	return newSpec(doc), nil
}
