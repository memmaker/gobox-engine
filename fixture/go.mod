module gobox/fixture

go 1.26.2

require gobox/ecl v0.0.0

require (
	gobox/asset v0.0.0
	gobox/dax v0.0.0 // indirect
	gobox/ega v0.0.0 // indirect
	gobox/game v0.0.0 // indirect
)

replace gobox/ecl => ../ecl

replace gobox/dax => ../dax

replace gobox/ega => ../ega

replace gobox/game => ../game

replace gobox/asset => ../asset

require gobox/importer v0.0.0

replace gobox/importer => ../importer

require gobox/core v0.0.0 // indirect

replace gobox/core => ../core
