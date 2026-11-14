package gpu

import "testing"

// TestMeshSetReleasesOwn: Set lets go of the mesh's own buffer (no GC
// release), and Dispose then leaves the frame's stream alone.
func TestMeshSetReleasesOwn(t *testing.T) {
	m := NewMesh(nil)
	if !m.own {
		t.Fatal("a new mesh does not own its buffer")
	}
	m.Set(make([]MeshVertex, 3))
	if m.own {
		t.Fatal("Set kept the mesh's own buffer")
	}
	m.Dispose() // the stream's: nothing released
	if m.buf == nil {
		t.Fatal("Dispose released the stream's buffer")
	}
}
