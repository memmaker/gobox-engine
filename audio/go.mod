module gobox/audio

go 1.26.2

require (
	github.com/ebitengine/oto/v3 v3.5.0
	github.com/jfreymuth/oggvorbis v1.0.5
)

require github.com/jfreymuth/vorbis v1.0.2 // indirect

require (
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	gobox/amiga v0.0.0
	golang.org/x/sys v0.47.0 // indirect
)

replace gobox/amiga => ../amiga
