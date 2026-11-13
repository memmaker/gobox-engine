package app

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CopyFS writes the files under root in fsys into dst (root's own path cut
// off), overwriting what is there: a game's -setup step puts its own
// embedded assets beside the imported game files.
func CopyFS(fsys fs.FS, root, dst string) error {
	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, filepath.FromSlash(strings.TrimPrefix(p, path.Clean(root)+"/")))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}
