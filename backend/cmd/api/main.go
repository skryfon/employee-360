// Command api is the Employee360 API server entrypoint.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the binary exists
// so `go build ./...` and `make dev` succeed, but config loading, the
// database connection, the middleware chain, routes, and graceful
// shutdown are not wired yet. They land in A2 (config, DB connectivity
// & migration runner) and A3 (middleware chain, health route & server
// bootstrap).
package main

import "fmt"

func main() {
	fmt.Println("employee360 api: scaffolding only (Cycle 1 / A1) — server bootstrap lands in A3")
}
