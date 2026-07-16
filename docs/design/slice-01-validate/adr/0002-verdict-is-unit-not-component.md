# Verdict codes are units, not component scenarios

The four incompatibility codes (`OP_NOT_IN_PROVIDER`, `MISSING_REQUIRED_REQUEST_FIELD`,
`READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`; exit 1) are a **domain verdict** produced on the success
path by `CompareOperation`/`DeriveProviderOperation` — `incompatible` is a legitimate answer, not a
pipe error — so they are proven by **unit** tests (the R1–R4 boundaries), and the component-scenario
count is `1 (happy) + Σ adapter branches = 6` (only the config/io `error.code`s at exit 2/3). A future
reader seeing 9 Cockburn Extensions but 6 component scenarios would otherwise wonder why; the
alternative (one component scenario per Extension) was rejected because it re-proves business logic the
units own and inflates the black-box suite. This fixes the test taxonomy and the coverage gate, so it
is recorded.
