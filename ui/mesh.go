package ui

// Tile meshes: a tile named <name> may have <name>.gltf (and its PNG) in the
// series' combat/meshes. Model units are 16 to a board square, origin at the
// square's centre on the floor, +Y up, front toward +Z.

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"io/fs"
	"math"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"gobox/gpu"
)

// meshUnits is a model's units to a board square.
const meshUnits = 16

type mat [16]float64 // column-major, as glTF

func (a mat) mul(b mat) (c mat) {
	for col := range 4 {
		for row := range 4 {
			for k := range 4 {
				c[col*4+row] += a[k*4+row] * b[col*4+k]
			}
		}
	}
	return c
}

// local is a node's own transform: its matrix, else translation * rotation * scale.
func local(n *gltf.Node) mat {
	if n.Matrix != gltf.DefaultMatrix && n.Matrix != [16]float64{} {
		return n.Matrix
	}
	t, q, s := n.TranslationOrDefault(), n.RotationOrDefault(), n.ScaleOrDefault()
	x, y, z, w := q[0], q[1], q[2], q[3]
	return mat{
		(1 - 2*(y*y+z*z)) * s[0], 2 * (x*y + z*w) * s[0], 2 * (x*z - y*w) * s[0], 0,
		2 * (x*y - z*w) * s[1], (1 - 2*(x*x+z*z)) * s[1], 2 * (y*z + x*w) * s[1], 0,
		2 * (x*z + y*w) * s[2], 2 * (y*z - x*w) * s[2], (1 - 2*(x*x+y*y)) * s[2], 0,
		t[0], t[1], t[2], 1,
	}
}

// readMesh reads name.gltf from fsys: its triangles in model units (nodes'
// transforms applied) and the first textured primitive's picture (nil: none).
func readMesh(fsys fs.FS, name string) ([]gpu.MeshVertex, image.Image, error) {
	b, err := fs.ReadFile(fsys, name+".gltf")
	if err != nil {
		return nil, nil, err
	}
	doc := new(gltf.Document)
	if err := gltf.NewDecoderFS(bytes.NewReader(b), fsys).Decode(doc); err != nil {
		return nil, nil, err
	}
	var vs []gpu.MeshVertex
	var img image.Image
	var walk func(i int, parent mat) error
	walk = func(i int, parent mat) error {
		n := doc.Nodes[i]
		m := parent.mul(local(n))
		if n.Mesh != nil {
			for _, p := range doc.Meshes[*n.Mesh].Primitives {
				var err error
				if vs, err = appendPrimitive(vs, doc, p, m); err != nil {
					return err
				}
				if img == nil && p.Material != nil {
					if t := doc.Materials[*p.Material].PBRMetallicRoughness; t != nil && t.BaseColorTexture != nil {
						if img, err = readTexture(fsys, doc, t.BaseColorTexture.Index); err != nil {
							return err
						}
					}
				}
			}
		}
		for _, c := range n.Children {
			if err := walk(c, m); err != nil {
				return err
			}
		}
		return nil
	}
	var roots []int
	if len(doc.Scenes) > 0 {
		sc := 0
		if doc.Scene != nil {
			sc = *doc.Scene
		}
		roots = doc.Scenes[sc].Nodes
	} else {
		for i := range doc.Nodes {
			roots = append(roots, i)
		}
	}
	for _, r := range roots {
		if err := walk(r, gltf.DefaultMatrix); err != nil {
			return nil, nil, err
		}
	}
	if len(vs) == 0 {
		return nil, nil, errors.New("no triangles")
	}
	return vs, img, nil
}

func appendPrimitive(vs []gpu.MeshVertex, doc *gltf.Document, p *gltf.Primitive, m mat) ([]gpu.MeshVertex, error) {
	if p.Mode != gltf.PrimitiveTriangles {
		return vs, errors.New("only triangle lists")
	}
	at, ok := p.Attributes[gltf.POSITION]
	if !ok {
		return vs, errors.New("no positions")
	}
	pos, err := modeler.ReadPosition(doc, doc.Accessors[at], nil)
	if err != nil {
		return vs, err
	}
	var uv [][2]float32
	if a, ok := p.Attributes[gltf.TEXCOORD_0]; ok {
		if uv, err = modeler.ReadTextureCoord(doc, doc.Accessors[a], nil); err != nil {
			return vs, err
		}
	}
	var nor [][3]float32
	if a, ok := p.Attributes[gltf.NORMAL]; ok {
		if nor, err = modeler.ReadNormal(doc, doc.Accessors[a], nil); err != nil {
			return vs, err
		}
	}
	var idx []uint32
	if p.Indices != nil {
		if idx, err = modeler.ReadIndices(doc, doc.Accessors[*p.Indices], nil); err != nil {
			return vs, err
		}
	} else {
		for i := range pos {
			idx = append(idx, uint32(i))
		}
	}
	for _, i := range idx {
		if int(i) >= len(pos) {
			return vs, errors.New("index out of range")
		}
		v := gpu.MeshVertex{R: 1, G: 1, B: 1, A: 1, NY: 1}
		P := pos[i]
		x, y, z := float64(P[0]), float64(P[1]), float64(P[2])
		v.X = float32(m[0]*x + m[4]*y + m[8]*z + m[12])
		v.Y = float32(m[1]*x + m[5]*y + m[9]*z + m[13])
		v.Z = float32(m[2]*x + m[6]*y + m[10]*z + m[14])
		if int(i) < len(uv) {
			v.U, v.V = uv[i][0], uv[i][1]
		}
		if int(i) < len(nor) {
			nx, ny, nz := float64(nor[i][0]), float64(nor[i][1]), float64(nor[i][2])
			// ponytail: the matrix itself, not its inverse transpose: exact for rotations and uniform scales
			ax, ay, az := m[0]*nx+m[4]*ny+m[8]*nz, m[1]*nx+m[5]*ny+m[9]*nz, m[2]*nx+m[6]*ny+m[10]*nz
			if l := math.Sqrt(ax*ax + ay*ay + az*az); l > 0 {
				v.NX, v.NY, v.NZ = float32(ax/l), float32(ay/l), float32(az/l)
			}
		}
		vs = append(vs, v)
	}
	return vs, nil
}

// readTexture is the PNG of texture i: a file beside the model, a data URI or a buffer view.
func readTexture(fsys fs.FS, doc *gltf.Document, i int) (image.Image, error) {
	if i >= len(doc.Textures) || doc.Textures[i].Source == nil {
		return nil, errors.New("texture without a source")
	}
	im := doc.Images[*doc.Textures[i].Source]
	var b []byte
	var err error
	switch {
	case im.BufferView != nil:
		b, err = modeler.ReadBufferView(doc, doc.BufferViews[*im.BufferView])
	case strings.HasPrefix(im.URI, "data:"):
		_, data, _ := strings.Cut(im.URI, ",")
		b, err = base64.StdEncoding.DecodeString(data)
	default:
		b, err = fs.ReadFile(fsys, im.URI)
	}
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(b))
}
