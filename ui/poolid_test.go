package ui

import "testing"

func TestPoolID(t *testing.T) {
	if poolID("portraits", "head_003", "body_017") == nil || poolID("combat", "combat_head_001", "combat_body_001") == nil {
		t.Fatal("pool id not drawn")
	}
	if poolID("portraits", "", "") != nil || poolID("portraits", "head_003", "") != nil {
		t.Fatal("bad id drawn")
	}
}
