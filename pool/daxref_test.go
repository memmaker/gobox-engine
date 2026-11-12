package pool

import (
	"path/filepath"
	"sync"

	"gobox/dax"
)

// refLoaders are the tests' reference decoders by data directory: the
// engine reads the imported files, a test comparing with the originals
// opens the archives itself (refDAX).
var refLoaders sync.Map

// refDAX is the DOS archives of e's data directory with the Amiga ones
// standing in, as the import reads them.
func refDAX(e *Engine) *dax.Loader {
	if l, ok := refLoaders.Load(e.data); ok {
		return l.(*dax.Loader)
	}
	l, err := dax.NewLoader(filepath.Join(e.data, DOSDir))
	if err != nil {
		panic(err) // the tests that call it have the game files
	}
	l.UseAmiga(filepath.Join(e.data, AmigaDir))
	actual, _ := refLoaders.LoadOrStore(e.data, l)
	return actual.(*dax.Loader)
}
