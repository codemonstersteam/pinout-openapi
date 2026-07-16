# SpecLoader unifies spec_path and spec_url behind one io: http object

The provider spec is reachable via `spec_path` (local file) XOR `spec_url` (HTTP GET). We hide both
behind **one** `SpecLoader` object tagged `io: http` — with file-vs-url an internal strategy — rather
than splitting them into a filesystem object + an HTTP object plus a selector, because `kin-openapi`'s
loader already unifies file/URI acquisition + `$ref` resolution behind one API, and the network path
(timeout, Bearer token, `HTTP_ERROR`/`TIMEOUT_ERROR` budgets) is the constrained one that must be
designed with the `http-io` skill; keeping it one seam avoids two objects sharing the same `PARSE_ERROR`
branch. The cost of reversing later (splitting the seam) is a module-boundary change, hence recorded.
