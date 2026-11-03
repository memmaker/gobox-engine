# gobox/gpu API (the slice of Ebiten por and Blobber use; gogpu/wgpu, pure Go)

Names mirror ebiten: swap the import. Subpackage gobox/gpu/vector draws filled and stroked paths.

## Loop
- `RunGame(Game) error`: opens a resizable window, blocks. `Game{Update() error; Draw(*Image); Layout(w,h)->(sw,sh)}`.
  Update runs at TPS (default 60, `SetTPS`), Draw once a frame into a screen of Layout's size. Return `ErrTermination` from Update to quit cleanly.
- Window: `SetWindowTitle/Size/Aspect(w,h)`, `SetFullscreen`/`IsFullscreen`, `DeviceScaleFactor`, `SetCursorCaptured`/`IsCursorCaptured` (mouselook: read `MouseDelta`).
- `Sized(&img,w,h) bool`: (re)make an image of that size, true if new. `Stats` (draw/pass counters), `Skip(name)` (POR_SKIP).

## Input (polled in Update; state is per tick)
- Keys: `IsKeyPressed`, `IsKeyJustPressed`, `KeyPressDuration` (ticks), `AppendPressedKeys`, `AppendJustPressedKeys`. `Key*` consts, plus `KeyShift/Control/Alt/Meta` meaning either side.
- Text: `AppendInputChars([]rune)` the runes typed since the last tick.
- Mouse: `CursorPosition()` (screen-image pixels), `MouseDelta`, `Wheel`, `IsMouseButtonPressed/JustPressed(MouseButtonLeft|Right|Middle)`, `AppendDroppedFiles([]Drop)`.
- Pad: `AppendGamepadIDs`, `IsStandardGamepadLayoutAvailable(id)`, `IsStandardGamepadButtonPressed`, `StandardGamepadButtonPressDuration`, `StandardGamepadAxisValue`. Standard layout only (named by place, e.g. RightBottom = A).
  Backends: GameController (darwin), XInput (windows), other builds per pad_other.go.

## Images
- `NewImage(w,h)`, `NewImageFromImage(image.Image)`; methods `Bounds, Size, At, SubImage, Clear, Fill, Deallocate, ReadPixels, WritePixels`.
- `ReadPixels(px)` / `WritePixels(px)`: premultiplied RGBA, Dx*Dy*4 bytes. `At` reads back too (slow: tests only).
- `DrawImage(src, *DrawImageOptions{GeoM, ColorScale, Filter, Blend})`; `GeoM` has Translate/Scale/Concat/Apply; `ColorScale` Scale/ScaleAlpha (zero value = no change).
- `DrawTriangles(vs []Vertex, is []uint16, img, *DrawTrianglesOptions{Filter,Blend})`; img nil = white. Vertex: DstX/Y, SrcX/Y (source pixels), ColorR..A, Custom0..3. Indices are uint16 (65535 vertices per call).
- `Filter`: FilterNearest, FilterLinear. `Blend`: BlendSourceOver, BlendLighter (additive), BlendSourceOverBack (depth tested, not written).

## Shaders (WGSL in the role of Kage; pixel units)
- `NewShader([]byte)`: declares `struct Uniforms{...}` and `fn Fragment(dstPos, srcPos, color, custom) -> vec4f`; helpers `imageSrcNAt(pos)`, `imageSrc0Origin/Size()`, `imageDstOrigin()`. Needs no window to compile.
- `DrawRectShader(w,h,*Shader,*DrawRectShaderOptions{GeoM,Uniforms map[string]any,Images [4]*Image,Blend})`, `DrawTrianglesShader(vs,is,s,*DrawTrianglesShaderOptions)`.

## 3D meshes (depth tested)
- `Mat4` (column-major, WebGPU clip z 0..1): `Ident`, `Mul`. `MeshVertex{X,Y,Z,U,V,NX,NY,NZ,R,G,B,A}`.
- `NewMesh(vs)`, `Set(vs)` (re-upload each frame before drawing), `Sub(first,n)`, `Dispose`.
- `DrawMesh(m, tex, viewProj, *MeshOptions{Model,Light,Ambient,Tint,Eye,Spec,Point,PointColor,Blend})`; `MeshBlend`: MeshOpaque (alpha tested, lit), MeshAlpha, MeshAdd (unlit, no depth write).
- `ClearDepth()` before a new set of meshes on an image. `NewMeshShader(src)` + `DrawMeshShader(m,s,imgs,uniforms []float32,blend)` for custom vs/fs (uniform block U starts with `rect: array<vec4f,4>`).

## How a game feeds it
Implement Game, call RunGame once. Per Update poll input; per Draw draw into `screen` only. Images must be made after the window runs (RunGame has started), except NewShader.

## Gotchas
- Global state: one window/device/input struct per process (package vars `win`, `dev`, `in`, `batch`); one RunGame only.
- Env switches read at runtime: POR_FPS, POR_SHOT, POR_SKIP, POR_PASSES, POR_NOVSYNC, POR_OFFSCREEN, POR_GLITCH.
- Draws are batched; ReadPixels/At flush. Deallocated images must not be used again.
- `SubImage` keeps the parent's coordinates; DrawImage draws src's bounds at the origin.
- Mesh.Set data lives in the frame's vertex stream: Set each frame, not once.
