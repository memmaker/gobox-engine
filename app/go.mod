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

require (
	gobox/audio v0.0.0
	gobox/dax v0.0.0
)

require (
	gobox/asset v0.0.0
	gobox/game v0.0.0 // indirect
)

replace gobox/dax => ../dax

replace gobox/audio => ../audio

replace gobox/amiga => ../amiga

replace gobox/game => ../game

replace gobox/geo => ../geo

replace gobox/combat => ../combat

replace gobox/walls => ../walls

replace gobox/sprites => ../sprites

replace gobox/asset => ../asset

replace gobox/core => ../core

require (
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/memmaker/opl2 v0.1.0 // indirect
	gobox/amiga v0.0.0 // indirect
	gobox/core v0.0.0 // indirect
)
