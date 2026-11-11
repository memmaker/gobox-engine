module gobox/sprites/cmd/poolsprites

go 1.26.2

require (
	gobox/dax v0.0.0
	gobox/ega v0.0.0
)

replace gobox/dax => ../../../dax

replace gobox/ega => ../../../ega

require gobox/asset v0.0.0 // indirect

replace gobox/asset => ../../../asset
