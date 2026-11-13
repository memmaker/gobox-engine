module pod

go 1.26.2

require (
	gobox/app v0.0.0
	gobox/char v0.0.0
	gobox/core v0.0.0
	gobox/dax v0.0.0
	gobox/ecl v0.0.0
	gobox/exe v0.0.0
	gobox/front v0.0.0
	gobox/game v0.0.0
	gobox/geo v0.0.0
	gobox/rules v0.0.0
)

require (
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/qmuntal/gltf v0.29.0 // indirect
	gobox/adf v0.0.0 // indirect
	gobox/ega v0.0.0 // indirect
	gobox/ui v0.0.0 // indirect
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
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/memmaker/opl2 v0.1.0 // indirect
	gobox/amiga v0.0.0 // indirect
	gobox/amigaimport v0.0.0
	gobox/asset v0.0.0
	gobox/audio v0.0.0
	gobox/blobber v0.0.0 // indirect
	gobox/combat v0.0.0 // indirect
	gobox/gpu v0.0.0 // indirect
	gobox/importer v0.0.0
	gobox/input v0.0.0 // indirect
	gobox/sprites v0.0.0 // indirect
	gobox/walls v0.0.0
	golang.org/x/image v0.47.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace (
	github.com/go-webgpu/goffi => ../../../goffi
	github.com/gogpu/gogpu => ../../../goplay/third_party/gogpu
	github.com/gogpu/wgpu => ../../../gogpu-wgpu
	github.com/retro-ware/gameimport => ../../../gameimport
	gobox/adf => ../../adf
	gobox/amiga => ../../amiga
	gobox/app => ../../app
	gobox/audio => ../../audio
	gobox/blobber => ../../blobber
	gobox/char => ../../char
	gobox/combat => ../../combat
	gobox/dax => ../../dax
	gobox/ecl => ../../ecl
	gobox/ega => ../../ega
	gobox/exe => ../../exe
	gobox/fixture => ../../fixture
	gobox/front => ../../front
	gobox/game => ../../game
	gobox/geo => ../../geo
	gobox/gpu => ../../gpu
	gobox/input => ../../input
	gobox/rules => ../../rules
	gobox/ui => ../../ui
	gobox/walls => ../../walls
)

replace gobox/sprites => ../../sprites

replace gobox/amigaimport => ../../amigaimport

replace gobox/asset => ../../asset

replace gobox/importer => ../../importer

replace gobox/core => ../../core

require (
	gobox/creation v0.0.0
	rules1e v0.0.0
)

replace gobox/creation => ../../creation

replace rules1e => ../../rules1e

require gobox/pool v0.0.0

replace gobox/pool => ../../pool
