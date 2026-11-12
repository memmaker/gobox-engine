module gobox/pool

go 1.26.2

require github.com/jfreymuth/vorbis v1.0.2 // indirect

require (
	gobox/core v0.0.0
	gobox/creation v0.0.0
	por v0.0.0
	rules1e v0.0.0
)

require (
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/memmaker/opl2 v0.1.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

require (
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	gobox/amiga v0.0.0
	gobox/app v0.0.0
	gobox/asset v0.0.0
	gobox/audio v0.0.0
	gobox/char v0.0.0
	gobox/dax v0.0.0
	gobox/ecl v0.0.0
	gobox/ega v0.0.0
	gobox/game v0.0.0
	gobox/geo v0.0.0
	gobox/importer v0.0.0
	gobox/rules v0.0.0
)

replace github.com/gogpu/wgpu => ../../gogpu-wgpu

replace github.com/go-webgpu/goffi => ../../goffi

replace github.com/gogpu/gogpu => ../../goplay/third_party/gogpu // macOS fullscreen fixes, see goplay/third_party/README.md

replace gobox/core => ../core

replace rules1e => ../../rules1e

replace gobox/gpu => ../gpu

replace gobox/input => ../input

replace gobox/blobber => ../blobber

replace gobox/combat => ../combat

replace gobox/adf => ../adf

replace gobox/audio => ../audio

replace gobox/amiga => ../amiga

replace gobox/app => ../app

replace gobox/dax => ../dax

replace gobox/importer => ../importer

replace gobox/ega => ../ega

replace gobox/exe => ../exe

replace gobox/fixture => ../fixture

replace gobox/geo => ../geo

replace gobox/ui => ../ui

replace gobox/walls => ../walls

replace gobox/front => ../front

replace gobox/rules => ../rules

replace gobox/char => ../char

replace gobox/amigaimport => ../amigaimport

replace gobox/ecl => ../ecl

replace gobox/game => ../game

replace gobox/sprites => ../sprites

replace gobox/asset => ../asset

replace gobox/creation => ../creation

replace por => ../../por-go
