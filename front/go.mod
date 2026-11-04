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

require gobox/audio v0.0.0

require github.com/memmaker/opl2 v0.1.0 // indirect

require (
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	gobox/amiga v0.0.0
	gobox/game v0.0.0
)

replace gobox/audio => ../audio

replace gobox/amiga => ../amiga

replace gobox/game => ../game
