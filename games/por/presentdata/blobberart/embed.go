// Package blobberart holds the remastered game's wall table and maps (the wall images are gobox/walls) (they
// came from gobox-engine/blobber's old embedded assets).
package blobberart

import "embed"

//go:embed assets/walls/walls.txt maps/*.map
var FS embed.FS
