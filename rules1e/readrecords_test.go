package rules1e

import "testing"

// A zero record size once looped forever, growing memory without bound.
func TestReadRecordsZeroSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("ReadRecords with size 0 did not panic")
		}
	}()
	ReadRecords[Item]([]byte{1, 2, 3}, 0)
}
