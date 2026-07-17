// Package contract — I/O pipe: read + parse the E-harness consumed-contract artifact
// (contracts.md §ContractStore.Load, module-tree.md). Pure pipe, no transformation:
// types arrive already typed (FRD Data-dictionary B) — no type inference here.
package contract

import (
	"encoding/json"
	"os"

	"pinout-openapi/internal/validate/domain"
)

// Store — the ContractStore port: reads a consumed-contract JSON file by path.
type Store struct{}

// NewStore — constructor of the file-based ContractStore.
func NewStore() Store { return Store{} }

// Load reads and parses the consumed-contract artifact at path.
// path non-empty is the caller's antecedent (Config-validated); Load itself is a pipe
// and does not re-validate it beyond what os.ReadFile needs.
func (Store) Load(path string) (domain.ConsumedContract, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return domain.ConsumedContract{}, domain.ErrFileNotFound
	}

	var out domain.ConsumedContract
	if err := json.Unmarshal(raw, &out); err != nil {
		return domain.ConsumedContract{}, domain.ErrParse
	}

	return out, nil
}
