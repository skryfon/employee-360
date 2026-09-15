// Command bootstrap will seed the system tenant, core roles, and the
// platform Super Admin (internal/infrastructure/database/seeder).
//
// This is Cycle 1 (project scaffolding, ticket A1) — there are no
// Tenant/Role/User entities or migrations yet, and database
// connectivity lands in A2. Seeding itself lands in a later cycle once
// those exist.
package main

import "fmt"

func main() {
	fmt.Println("employee360 bootstrap: scaffolding only (Cycle 1 / A1) — nothing to seed yet")
}
