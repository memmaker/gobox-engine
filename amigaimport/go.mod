module gobox/amigaimport

go 1.26.2

require (
	gobox/adf v0.0.0
	gobox/amiga v0.0.0
	gobox/dax v0.0.0
)

require (
	gobox/asset v0.0.0
	gobox/ega v0.0.0 // indirect
)

replace (
	gobox/adf => ../adf
	gobox/amiga => ../amiga
	gobox/dax => ../dax
	gobox/ega => ../ega
)

replace gobox/asset => ../asset
