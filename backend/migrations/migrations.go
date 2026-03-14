// Package migrations holds the embedded SQL migration files for goose.
// Keeping migrations at the backend root makes them easy to find and edit.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
