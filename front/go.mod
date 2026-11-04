module gobox/front

go 1.26.2

require (
	gobox/gpu v0.0.0
	golang.org/x/image v0.47.0
)

require (
	github.com/go-webgpu/goffi v0.6.3 // indirect
	github.com/go-webgpu/webgpu v0.5.5 // indirect
	github.com/gogpu/gogpu v0.54.0 // indirect
	github.com/gogpu/gpucontext v0.31.3 // indirect
	github.com/gogpu/gputypes v0.8.0 // indirect
	github.com/gogpu/naga v0.19.0 // indirect
	github.com/gogpu/wgpu v0.34.5 // indirect
	gobox/blobber v0.0.0
	gobox/combat v0.0.0
	gobox/dax v0.0.0
	gobox/ega v0.0.0
	gobox/input v0.0.0
	gobox/ui v0.0.0
	gobox/walls v0.0.0
	golang.org/x/sys v0.49.0 // indirect
)

replace gobox/gpu => ../gpu

replace gobox/blobber => ../blobber

replace gobox/combat => ../combat

replace gobox/ega => ../ega

replace gobox/input => ../input

replace gobox/ui => ../ui

replace gobox/walls => ../walls

replace gobox/dax => ../dax
