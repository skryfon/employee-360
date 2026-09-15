// Command migrate will apply or revert golang-migrate SQL migrations
// from migrations/ against the configured database.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the binary exists
// so `go build ./...` and `make migrate` succeed, but the actual
// migration runner (config + golang-migrate wiring) lands in A2
// (config, DB connectivity & migration runner).
package main

import "fmt"

func main() {
	fmt.Println("employee360 migrate: scaffolding only (Cycle 1 / A1) — migration runner lands in A2")
}
