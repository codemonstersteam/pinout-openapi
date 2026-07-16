# Filesystem loaders are tagged io: none

`ConfigStore`, `ContractStore` and `ReportWriter` read/write **local files**, but the mandatory `io:`
enum is `none|http|llm|queue|db` — there is no `file` value and no filesystem io sub-skill to route to
(unlike `http`→`http-io`, `db`→`db-io`). We tag them **`io: none`** while keeping them isolated I/O
pipes: their distinguishable failure branches (`CONFIG_ERROR`, `FILE_NOT_FOUND`, `PARSE_ERROR`) are
still component-test adapter branches, and the `os` stdlib dependency is the "borderline-allowed"
category (like `clock.Clock`/`crypto/rand`), not one of the forbidden network/db/broker handles. The
alternative (inventing an off-enum `io: file`) would break the ticketer's io-router key. This is a
load-bearing io: classification, hence recorded.
