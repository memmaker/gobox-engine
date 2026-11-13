package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteJSON writes v as indented JSON to path, making its folder.
func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
