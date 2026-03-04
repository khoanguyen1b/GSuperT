package migrations

import "embed"

// Files contains all SQL migration files bundled into the binary.
//
//go:embed *.sql
var Files embed.FS
