package main

import (
	"testing"
)

func TestGet(t *testing.T) {
	lsm := Connect("files/2026-11-09.txt")
	value := lsm.Get("key2")
	if value != "value2" {
		t.Errorf("expected: value2, actual: %s", value)
		t.FailNow()
	}
}
