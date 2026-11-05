package combat

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"testing"
	"testing/fstest"
)

// The pilots stay inside their square: 16 units wide and deep, standing on the floor.
func TestReadMeshPilots(t *testing.T) {
	dir := os.DirFS("../sprites/series/buck-rogers/combat/meshes")
	for _, n := range []string{"hitek_tile_019", "hitek_tile_035", "hitek_tile_036", "hitek_tile_037", "hitek_tile_038", "hitek_tile_039", "hitek_tile_040", "hitek_tile_041", "hitek_tile_042", "hitek_tile_043", "hitek_tile_044", "lotek_tile_042"} {
		vs, img, err := readMesh(dir, n)
		if err != nil || img == nil || len(vs) == 0 || len(vs)%3 != 0 {
			t.Fatalf("%s: %d vertices, texture %v, %v", n, len(vs), img != nil, err)
		}
		for _, v := range vs {
			if math.Abs(float64(v.X)) > 8.001 || math.Abs(float64(v.Z)) > 8.001 || v.Y < -0.001 {
				t.Fatalf("%s: vertex (%v %v %v) outside the square", n, v.X, v.Y, v.Z)
			}
		}
	}
}

// A node's translation and scale move its mesh.
func TestReadMeshNodeTransform(t *testing.T) {
	var buf []byte
	for _, f := range []float32{0, 0, 0, 1, 0, 0, 0, 1, 0} {
		buf = binary.LittleEndian.AppendUint32(buf, math.Float32bits(f))
	}
	doc := fmt.Sprintf(`{"asset":{"version":"2.0"},"scene":0,"scenes":[{"nodes":[0]}],
	"nodes":[{"children":[1],"translation":[16,0,0]},{"mesh":0,"scale":[2,2,2]}],
	"meshes":[{"primitives":[{"attributes":{"POSITION":0}}]}],
	"accessors":[{"bufferView":0,"componentType":5126,"count":3,"type":"VEC3","min":[0,0,0],"max":[1,1,0]}],
	"bufferViews":[{"buffer":0,"byteLength":36}],
	"buffers":[{"byteLength":36,"uri":"data:application/octet-stream;base64,%s"}]}`, base64.StdEncoding.EncodeToString(buf))
	vs, _, err := readMesh(fstest.MapFS{"m.gltf": {Data: []byte(doc)}}, "m")
	if err != nil || len(vs) != 3 {
		t.Fatalf("%d vertices, %v", len(vs), err)
	}
	if vs[1].X != 18 || vs[2].Y != 2 {
		t.Errorf("got (%v %v) and (%v %v), want x 18 and y 2", vs[1].X, vs[1].Y, vs[2].X, vs[2].Y)
	}
}
