package compatcheck

import (
	"time"

	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// NewDeps assembles the concrete Config/Consumer/Provider adapters built by tickets 03 (config)
// and 04 (spec) into three of the four ports declared in domain.go (ticket 06).
//
// DEVIATION (flagged): the fourth port, Report, is deliberately left unset (zero value) here —
// wiring it in this file would create a Go import cycle. report.Writer (ticket 07) already
// imports this package's root (compatcheck) for the Outcome return type and the
// ErrReportWrite sentinel (report/writer.go); if this file — package compatcheck, same package
// as domain.go/head.go — also imported report/ to call report.NewWriter(), the import graph
// would close: compatcheck -> report -> compatcheck, which `go build` rejects outright. Neither
// side can be changed without touching an already-green module (report/, ticket 07) or the
// port declarations (domain.go, ticket 06) — out of this ticket's scope (register.go +
// cmd/app/main.go only). cmd/app/main.go is package main, imported by nothing, so it can safely
// import both compatcheck and report; it completes Deps.Report = report.NewWriter() right after
// calling NewDeps(). This is the composition root's only viable split given the existing
// (unchanged) package graph.
//
// Deps.Provider is a ProviderSpecReaderFactory, not a ready spec.ProviderSpecReader instance
// (see domain.go's ProviderSpecReaderFactory doc for the two-phase-timeout rationale): the
// timeout is only known once config.NewConfig has run, deep inside ProcessCompatCheck's pipe
// (head.go step 5), while NewDeps runs once at process start-up before any config is read. The
// closure below defers spec.NewProviderSpecReader's construction to that later point.
func NewDeps() Deps {
	return Deps{
		Config:   config.NewConfigReader(),
		Consumer: spec.NewConsumerSpecReader(),
		Provider: func(timeout time.Duration) ProviderSpecReader {
			return spec.NewProviderSpecReader(timeout)
		},
	}
}
