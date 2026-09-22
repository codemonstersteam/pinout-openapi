# CHORE-PLAN — 001-report-canon-doc

**Mode:** chore (`.agent/planner/mode` = `chore`). **Source BRD:**
[`.agent/planner/brd.md`](../../../.agent/planner/brd.md) (0 open questions, decisions D1–D8 resolved).
**Source BR tickets:** [`debt/01-report-canon-doc.md`](../../../debt/01-report-canon-doc.md) (this chore),
[`debt/02-report-schema-1.1.md`](../../../debt/02-report-schema-1.1.md) (ticket 2/2, out of scope — a
separate future run). No FRD, no design package, no tickets folder: this is a single agent-ready doc-writing
pass, not a slice.

## Scope

Author exactly one new file, `docs/report-format.md` — the ecosystem-wide canonical specification of the
validator report format, version 1.1 — content dictated verbatim by the BRD's Data dictionary A (schema
fields), Data dictionary B (11 mandatory sections B1–B11), the failure-mode map, and NFR N1–N12. No code,
no schema-file edit, no test. The BRD is the sole content source; this plan indexes it and does not restate
its tables.

## Files to touch

| path | action | note |
|---|---|---|
| `docs/report-format.md` | **create** | the deliverable; content = BRD Data dictionary A + B (B1–B11) + failure-mode map, verbatim per fit criteria |
| `.agent/planner/chore-dir` | write pointer | `docs/chores/001-report-canon-doc` (this run) |

No other file changes are in scope. `api-specification/`, `internal/`, tests, and fixtures are explicitly
**out of bounds** (BRD "MUST NOT touch"; N8) — they belong to ticket 2/2
(`debt/02-report-schema-1.1.md`), a separate run.

## Verification command

Mechanical checks, run from the repo root after the file is written. Exits non-zero on the first failure.

```bash
#!/usr/bin/env bash
set -euo pipefail

DOC="docs/report-format.md"
SYNC_SCHEMA="api-specification/report.schema.json"
ASYNC_SCHEMA="../pinout-asyncapi/api-specification/report.schema.json"

fail() { echo "FAIL: $1"; exit 1; }

# N1 — file exists at the canonical path
[ -f "$DOC" ] || fail "N1: $DOC missing"

# N8 — scope containment: nothing changed outside docs/report-format.md
# (run against the base branch this chore is committed on top of, e.g. main)
BASE="${BASE_REF:-main}"
OUT_OF_SCOPE=$(git diff --name-only "$BASE"...HEAD -- . | grep -v '^docs/report-format\.md$' || true)
[ -z "$OUT_OF_SCOPE" ] || fail "N8: files changed outside docs/report-format.md: $OUT_OF_SCOPE"
git diff --name-only "$BASE"...HEAD -- api-specification/ internal/ | grep -q . && \
  fail "N8: api-specification/ or internal/ touched" || true

# N2 — schema_version 1.1 documented
grep -q '"1\.1"' "$DOC" || fail "N2: const \"1.1\" not documented"

# N12 — generated_at fit criterion (UTC, Z, second precision) stated verbatim
grep -qF '^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$' "$DOC" || fail "N12: generated_at regex not stated"

# N7 — enum fidelity vs the two frozen schemas (sync always local; async only if the
# sibling repo is checked out alongside this one — else this half is SKIPPED, not passed)
for code in OP_NOT_IN_PROVIDER MISSING_REQUIRED_REQUEST_FIELD READS_FIELD_NOT_PROVIDED TYPE_MISMATCH \
            PARSE_ERROR FILE_NOT_FOUND HTTP_ERROR TIMEOUT_ERROR CONFIG_ERROR; do
  grep -q "$code" "$DOC" || fail "N7: sync/shared code $code missing from doc"
done
if [ -f "$ASYNC_SCHEMA" ]; then
  for code in $(jq -r '.properties.errors.items.properties.code.enum[]' "$ASYNC_SCHEMA" | grep '^[A-Z_]*$'); do
    grep -q "$code" "$DOC" || fail "N7: async code $code missing from doc"
  done
else
  echo "SKIP N7-async: $ASYNC_SCHEMA not found (sibling repo not checked out)"
fi

# B1/N9/D4 — bold superseded metadata line present (not YAML front-matter)
grep -qE '^\*\*.*superseded.*docs/report-format\.md@2e0c232\^' "$DOC" || fail "B1/D4: superseded line missing or wrong form"
head -5 "$DOC" | grep -q '^---$' && fail "D4: YAML front-matter used instead of a bold metadata line"

# B11/N11/D8 — schema-file link + drift note present
grep -qF 'api-specification/report.schema.json' "$DOC" || fail "B11: schema-file link missing"
grep -qi 'reaches 1.1 in ticket 2/2' "$DOC" || fail "N11: drift note missing"

# N10/D7 — English (heuristic: no Cyrillic in the body)
LC_ALL=C grep -qP '[а-яА-ЯёЁ]' "$DOC" && fail "N10: Cyrillic characters found (doc must be English)" || true

echo "OK: mechanical checks passed"
```

```bash
# md-formatting skill checks (N4, N6) — run its Step 1–5 procedure against the file
# (heading blank lines, list/table/code-fence blank lines + language tags, link/anchor resolution)
```

### What this command does NOT prove (semantic, judgment-bound — reviewer's job at Gate #1)

| # | criterion | why not mechanical | who checks |
|---|---|---|---|
| N3 | B2 prose = B3 delta table, 0 disagreements | requires reading both sections for equivalence, not string match | plan-reviewer / fixer, side-by-side against BRD Data dictionary A |
| N5 | self-sufficiency — 0 facts a mirror repo still needs from `internal/` | requires judging completeness against BRD Data dictionary A + B5 | plan-reviewer / fixer |
| N9 | dead `verdicts[]`/`provider{}` shape appears only in the superseded note, never as normative form | needs context around each occurrence, not just presence | plan-reviewer / fixer |
| B4–B10 content correctness (exit-code grid, code dictionaries, invariant, provenance rationale, no-`verdicts[]` rationale, versioning policy, divergence note) | grep can confirm presence of keywords, not correctness of the stated rule | plan-reviewer / fixer, row-by-row against BRD Data dictionary B |

These map 1:1 to BRD UC-D1 Extensions 2a/3a/5a/7a/9a/9b and UC-D2 Extension 2a — the reviewer walks the
doc against those extensions plus the manual-check table above before Gate #1 sign-off.

## Rollback

Single-file, additive chore — no schema/code touched, so rollback is a plain revert:

1. `git rm docs/report-format.md` (or `git revert <commit-sha>` if already committed on trunk).
2. `rm -f .agent/planner/chore-dir` (or restore its prior state — this run's pointer only).
3. No migration, no data, no running service is affected — the file has no runtime reader yet.

## Traceability

| this plan | BRD anchor |
|---|---|
| file list | BRD "Deliverable shape and boundaries" |
| N1–N12 checks | BRD "NFR / measurable acceptance criteria" |
| B1–B11 manual-check rows | BRD "Data dictionary B — required sections of the document" |
| enum lists in the verification script | BRD "Data dictionary A" `errors[].code`; "Failure-mode map" |
| D1–D8 spot checks (D4 header form, D7 language, D8 drift note, D2/N12 regex) | BRD "Resolved decisions" table |

**Ticket count:** 0 — chore lane carries no FRD/design/tickets; this CHORE-PLAN is the sole handoff
artifact for Gate #1.
