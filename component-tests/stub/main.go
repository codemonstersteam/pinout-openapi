// Command provider-stub — real-protocol HTTP stub for slice-01-validate component scenarios
// 5 (HTTP_ERROR) and 6 (TIMEOUT_ERROR). It is a STUB, not an in-code mock (Fowler): a real
// HTTP server, reached by SUT over the real protocol via the compose service name
// "provider-stub". Two routes, one per adapter branch under test:
//
//	GET /error  -> 503 Service Unavailable (SpecLoader.Load's ErrHTTP branch)
//	GET /stall  -> sleeps past any reasonable settings.timeout, then answers 200
//	               (SpecLoader.Load's ErrTimeout branch; fixtures set settings.timeout=1)
package main

import (
	"log"
	"net/http"
	"time"
)

// stallDelay MUST exceed the bad-timeout fixture's settings.timeout (1s) by a wide margin
// so the fetch-side deadline fires deterministically, not by a race.
const stallDelay = 5 * time.Second

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/error", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
	})

	mux.HandleFunc("/stall", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(stallDelay):
		case <-r.Context().Done():
			// client (SUT) gave up first — expected once its own timeout fires.
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("openapi: 3.0.3\ninfo:\n  title: too-late\n  version: 1.0.0\npaths: {}\n"))
	})

	log.Println("provider-stub listening on :8080 (routes: /error, /stall)")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
