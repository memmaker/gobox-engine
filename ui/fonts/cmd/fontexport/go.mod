module gobox/ui/fonts/cmd/fontexport

go 1.26.2

require gobox/dax v0.0.0

require (
	gobox/asset v0.0.0 // indirect
	gobox/ega v0.0.0 // indirect
)

replace (
	gobox/dax => ../../../../dax
	gobox/ega => ../../../../ega
)

require gobox/ui v0.0.0

replace gobox/ui => ../../..

replace gobox/asset => ../../../../asset
