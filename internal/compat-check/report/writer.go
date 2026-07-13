// Package report is the fourth I/O adapter of slice compat-check (module-tree.md, contracts.md
// "ReportWriter"): a pure I/O pipe that persists the already-built compare.Report to the frozen
// api-specification/report.schema.json canon shape — no transformation of report data, only
// serialization + writing. Not unit-tested (module-tree.md) — proven by component scenarios
// CS1/CS2 (success path, report shape + byte-stability) and CS9 (failure path), via the fixer.
package report

import (
	"encoding/json"
	"fmt"
	"os"

	compatcheck "pinout-openapi/internal/compat-check"
	"pinout-openapi/internal/compat-check/compare"
	"pinout-openapi/internal/compat-check/config"
)

// jsonReport mirrors the frozen api-specification/report.schema.json shape exactly: field names,
// nesting and the required set. generated_at is never emitted — it is optional in the schema and
// excluded from the determinism guarantee (contracts.md "Determinism (CS1)"); omitting it keeps
// Write byte-stable by construction rather than by post-hoc filtering.
type jsonReport struct {
	Verdict    compare.Verdict `json:"verdict"`
	Consumer   jsonConsumer    `json:"consumer"`
	Provider   jsonProvider    `json:"provider"`
	Operations []jsonOperation `json:"operations"`
}

type jsonConsumer struct {
	Name string `json:"name"`
}

type jsonProvider struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type jsonOperation struct {
	Path     string          `json:"path"`
	Method   string          `json:"method"`
	Status   compare.Verdict `json:"status"`
	Findings []jsonFinding   `json:"findings"`
}

type jsonFinding struct {
	Rule     compare.Rule `json:"rule"`
	Location string       `json:"location"`
	Detail   string       `json:"detail"`
}

// Writer is the ReportWriter port's concrete adapter (internal/compat-check/domain.go's
// ReportWriter interface; contracts.md "ReportWriter" · io: file). No dependencies to hide — the
// filesystem path and stdout stream it touches are resolved per-call from Write's own arguments,
// so there is nothing to inject at construction (module-tree.md "autonomous I/O object").
type Writer struct{}

// NewWriter constructs the report Writer.
func NewWriter() Writer { return Writer{} }

// Write serializes report to the frozen report.schema.json canon shape and always prints the JSON
// body to stdout (contracts.md "the machine-readable stdout line (always, regardless of
// save_json_report)"); when settings.SaveJSONReport() is true it additionally persists the same
// bytes to settings.JSONReportFile(). Written for BOTH the compatible and incompatible verdict —
// incompatible is exit 1, not a Write error. Any failure (encode, file write, or stdout write)
// surfaces as compatcheck.ErrReportWrite, wrapped so errors.Is still matches
// (contracts.md "Error model": REPORT_WRITE_ERROR, exit 3).
func (w Writer) Write(r compare.Report, settings config.Settings) (compatcheck.Outcome, error) {
	body, err := json.MarshalIndent(toJSONReport(r), "", "  ")
	if err != nil {
		return compatcheck.Outcome{}, fmt.Errorf("%w: encode report: %v", compatcheck.ErrReportWrite, err)
	}

	if settings.SaveJSONReport() {
		if err := os.WriteFile(settings.JSONReportFile(), body, 0o644); err != nil {
			return compatcheck.Outcome{}, fmt.Errorf("%w: write %s: %v", compatcheck.ErrReportWrite, settings.JSONReportFile(), err)
		}
	}

	if _, err := fmt.Println(string(body)); err != nil {
		return compatcheck.Outcome{}, fmt.Errorf("%w: stdout: %v", compatcheck.ErrReportWrite, err)
	}

	return compatcheck.Outcome{Verdict: r.Verdict}, nil
}

// toJSONReport maps the domain compare.Report onto the frozen wire shape. Pipe only — no logic,
// no branching beyond the 1:1 field/slice walk.
func toJSONReport(r compare.Report) jsonReport {
	ops := make([]jsonOperation, 0, len(r.Operations))
	for _, op := range r.Operations {
		findings := make([]jsonFinding, 0, len(op.Findings))
		for _, f := range op.Findings {
			findings = append(findings, jsonFinding{
				Rule:     f.Rule,
				Location: f.Location,
				Detail:   f.Detail,
			})
		}
		ops = append(ops, jsonOperation{
			Path:     op.Path,
			Method:   op.Method,
			Status:   op.Status,
			Findings: findings,
		})
	}
	return jsonReport{
		Verdict:    r.Verdict,
		Consumer:   jsonConsumer{Name: r.ConsumerName},
		Provider:   jsonProvider{Name: r.ProviderName, Source: r.ProviderSource},
		Operations: ops,
	}
}
