module gobox/ui

go 1.26.2

require (
	gobox/gpu v0.0.0
	gobox/input v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-webgpu/goffi v0.6.3 // indirect
	github.com/go-webgpu/webgpu v0.5.5 // indirect
	github.com/gogpu/gogpu v0.54.0 // indirect
	github.com/gogpu/gpucontext v0.31.3 // indirect
	github.com/gogpu/gputypes v0.8.0 // indirect
	github.com/gogpu/naga v0.19.0 // indirect
	github.com/gogpu/wgpu v0.34.5 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace (
	gobox/gpu => ../gpu
	gobox/input => ../input
)
