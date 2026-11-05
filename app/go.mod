module gobox/app

go 1.26.2

require gobox/gpu v0.0.0

require (
	github.com/go-webgpu/goffi v0.6.3 // indirect
	github.com/go-webgpu/webgpu v0.5.5 // indirect
	github.com/gogpu/gogpu v0.54.0 // indirect
	github.com/gogpu/gpucontext v0.31.3 // indirect
	github.com/gogpu/gputypes v0.8.0 // indirect
	github.com/gogpu/naga v0.19.0 // indirect
	github.com/gogpu/wgpu v0.34.5 // indirect
	gobox/blobber v0.0.0
	gobox/ega v0.0.0
	gobox/front v0.0.0
	gobox/input v0.0.0
	gobox/ui v0.0.0
	golang.org/x/sys v0.49.0 // indirect
)

replace gobox/gpu => ../gpu

replace github.com/gogpu/wgpu => ../../gogpu-wgpu

replace github.com/go-webgpu/goffi => ../../goffi

replace github.com/gogpu/gogpu => ../../goplay/third_party/gogpu

replace gobox/blobber => ../blobber

replace gobox/front => ../front

replace gobox/input => ../input

replace gobox/ui => ../ui

replace gobox/ega => ../ega

require gobox/dax v0.0.0

require github.com/qmuntal/gltf v0.29.0 // indirect

replace gobox/dax => ../dax

replace gobox/audio => ../audio

replace gobox/amiga => ../amiga
