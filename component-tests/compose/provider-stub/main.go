// Command provider-stub — real-protocol HTTP double for the compat-check
// component tests' provider-side scenarios (ticket-02, slice-compat-check):
//
//	CS6 /unreachable — non-2xx (500) response, simulates PROVIDER_UNREACHABLE.
//	CS7 /timeout     — delays past any fixture's settings.timeout, simulates
//	                    PROVIDER_TIMEOUT.
//	CS8 /malformed    — 200 OK with a non-OpenAPI body, simulates PROVIDER_PARSE_ERROR.
//
// A stub, not an in-code mock (Fowler, "Mocks Aren't Stubs"): a real server
// answering over the real HTTP protocol, reached by the compose service name
// "provider-stub" — never a fake wired into the CLI's process.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// CS6 — PROVIDER_UNREACHABLE: non-2xx response.
	mux.HandleFunc("/unreachable", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "provider-stub: simulated 500 (unreachable)")
	})

	// CS7 — PROVIDER_TIMEOUT: reply well after any reasonable settings.timeout
	// (fixtures pin timeout=1s; 5s here is comfortably past that bound).
	mux.HandleFunc("/timeout", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "provider-stub: too slow, arrived after the client timeout")
	})

	// CS8 — PROVIDER_PARSE_ERROR: 200 OK, body is not a valid OpenAPI 3.x document.
	mux.HandleFunc("/malformed", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "this is not an OpenAPI document")
	})

	// /healthz — polled by the tester's own BeforeSuite readiness wait
	// (component-tests/steps/main_test.go); not a compat-check scenario itself.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	const addr = ":8080"
	log.Printf("provider-stub listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
