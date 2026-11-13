// Package pod holds the table spec tables.json: which data tables the
// importer reads out of the executable into data/rules.json (tablespec).
package pod

import _ "embed"

// TablesJSON is tables.json.
//
//go:embed tables.json
var TablesJSON []byte
