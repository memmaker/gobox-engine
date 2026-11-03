package adf

import "testing"

func TestReadRejectsShort(t *testing.T) {
	if _, _, err := Read(make([]byte, 100)); err == nil {
		t.Fatal("short image accepted")
	}
}
