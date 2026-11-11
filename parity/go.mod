module gobox/parity

go 1.26.2

require (
	gobox/blobber v0.0.0
	gobox/dax v0.0.0
	gobox/ecl v0.0.0
	gobox/ega v0.0.0
	gobox/front v0.0.0
	gobox/game v0.0.0
	gobox/gpu v0.0.0
	gobox/sprites v0.0.0
	gobox/walls v0.0.0
	por v0.0.0
)

require (
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/go-webgpu/goffi v0.6.3 // indirect
	github.com/go-webgpu/webgpu v0.5.5 // indirect
	github.com/gogpu/gogpu v0.54.0 // indirect
	github.com/gogpu/gpucontext v0.31.3 // indirect
	github.com/gogpu/gputypes v0.8.0 // indirect
	github.com/gogpu/naga v0.19.0 // indirect
	github.com/gogpu/wgpu v0.34.5 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/memmaker/opl2 v0.1.0 // indirect
	github.com/qmuntal/gltf v0.29.0 // indirect
	gobox/amiga v0.0.0 // indirect
	gobox/app v0.0.0 // indirect
	gobox/asset v0.0.0
	gobox/audio v0.0.0 // indirect
	gobox/char v0.0.0 // indirect
	gobox/combat v0.0.0 // indirect
	gobox/geo v0.0.0 // indirect
	gobox/input v0.0.0 // indirect
	gobox/rules v0.0.0 // indirect
	gobox/ui v0.0.0 // indirect
	golang.org/x/image v0.47.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	rules1e v0.0.0
)

// por's checkout beside gobox's (the parity tests read its presentdata goldens)
replace por => ../../por-go

replace gobox/gpu => ../gpu

replace gobox/input => ../input

replace gobox/blobber => ../blobber

replace gobox/combat => ../combat

replace gobox/adf => ../adf

replace gobox/audio => ../audio

replace gobox/amiga => ../amiga

replace gobox/app => ../app

replace gobox/dax => ../dax

replace gobox/ega => ../ega

replace gobox/exe => ../exe

replace gobox/fixture => ../fixture

replace gobox/geo => ../geo

replace gobox/ui => ../ui

replace gobox/walls => ../walls

replace gobox/front => ../front

replace gobox/sprites => ../sprites

replace gobox/rules => ../rules

replace gobox/char => ../char

replace gobox/amigaimport => ../amigaimport

replace gobox/game => ../game

replace github.com/gogpu/wgpu => ../../gogpu-wgpu

replace github.com/go-webgpu/goffi => ../../goffi

replace github.com/gogpu/gogpu => ../../goplay/third_party/gogpu

replace gobox/ecl => ../ecl

replace gobox/asset => ../asset

replace gobox/core => ../core

replace rules1e => ../../rules1e
