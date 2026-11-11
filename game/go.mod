module gobox/game

go 1.26.2

require (
	gobox/core v0.0.0
	gobox/importer v0.0.0
)

replace (
	gobox/core => ../core
	gobox/importer => ../importer
)
