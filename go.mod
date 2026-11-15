module gobox

go 1.26.2

require (
	buckrogers v0.0.0
	buckrogers2 v0.0.0
	github.com/ebitengine/oto/v3 v3.5.0
	github.com/go-webgpu/goffi v0.6.3
	github.com/gogpu/gogpu v0.54.0
	github.com/gogpu/gpucontext v0.31.3
	github.com/gogpu/gputypes v0.8.0
	github.com/gogpu/naga v0.19.0
	github.com/gogpu/wgpu v0.34.5
	github.com/hajimehoshi/go-mp3 v0.3.4
	github.com/jfreymuth/oggvorbis v1.0.5
	github.com/memmaker/opl2 v0.1.0
	github.com/qmuntal/gltf v0.29.0
	golang.org/x/image v0.47.0
)

require (
	ecl-br v0.0.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/go-webgpu/webgpu v0.5.5 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	golang.org/x/sys v0.49.0 // indirect
	rulesxxvc v0.0.0 // indirect
)


replace github.com/gogpu/wgpu => ../gogpu-wgpu

replace github.com/go-webgpu/goffi => ../goffi

// macOS fullscreen fixes, see goplay/third_party/README.md
replace github.com/gogpu/gogpu => ../goplay/third_party/gogpu

replace github.com/retro-ware/gameimport => ../gameimport

replace buckrogers => ../buckrogers-go

replace buckrogers2 => ../buckrogers2-go

replace ecl-br => ../buckrogers-core

replace rulesxxvc => ../rulesxxvc
