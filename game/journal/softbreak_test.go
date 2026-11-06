package journal

import (
	"testing"
	"testing/fstest"
)

// TestLoadJoinsBreaks: a word the printed journal broke ("men- ace") loads whole.
func TestLoadJoinsBreaks(t *testing.T) {
	b, err := Load(fstest.MapFS{"J.TXT": {Data: []byte("[1]\nhis men- ace, per- haps - no\n")}}, "J.TXT")
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Text("1")[0]; got != "his menace, perhaps - no" {
		t.Fatal(got)
	}
}
