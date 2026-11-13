package main

import "testing"

func TestOperands(t *testing.T) {
	if operands[41] != 14 || operands[1] != 1 {
		t.Fatal("operand table")
	}
}
