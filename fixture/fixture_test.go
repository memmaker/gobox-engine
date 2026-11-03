package fixture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRawSave(t *testing.T) {
	data, out := t.TempDir(), t.TempDir() // never the game's SAVE folder
	sv := make([]byte, svSize)
	sv[svCount] = 1
	sv[svNames] = 3
	copy(sv[svNames+1:], "BOB")
	os.WriteFile(filepath.Join(data, "BOB.SAV"), make([]byte, 0x200), 0o644)
	dec := filepath.Join(data, "d.txt")
	os.WriteFile(dec, []byte("# c\nwalk 1\n"), 0o644)
	g := Game{
		Root:          func(string) (string, error) { return data, nil },
		Open:          func(string, string) (State, error) { return OpenRawSave(sv) },
		ECL:           func(string, int) (int, []byte, error) { return 2, []byte{1, 2, 3}, nil },
		Member:        func(_, n, e string) ([]byte, error) { return os.ReadFile(filepath.Join(data, n+e)) },
		Strong:        MakeStrong,
		ParseDecision: func(l string) error { return nil },
	}
	err := Run(g, []string{"-out", out, "-block", "5", "-walls", "1,2,255", "-x", "7", "-y", "8", "-dir", "2",
		"-set", "4A05=1234", "-strong", "-decisions", dec})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(out, "SAVGAMA.DAT"))
	if b[svDisk] != 2 || b[svPos] != 7 || b[svPos+2] != 2 || b[svECL+1] != 2 {
		t.Fatalf("bad save %v", b[:4])
	}
	m, err := os.ReadFile(filepath.Join(out, "BOB.SAV"))
	if err != nil || m[0x32] != 250 {
		t.Fatal("member not strong", err)
	}
	if Run(g, []string{"-out", out, "-set", "zz=1"}) == nil {
		t.Fatal("bad -set accepted")
	}
}
