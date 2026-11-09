package main

import (
	"testing"
)

func TestGet(t *testing.T) {
	lsm := Connect("files/2026-11-09.txt")
	defer lsm.file.Close()
	value := lsm.Get("key2")
	if value != "value2" {
		t.Errorf("expected: value2, actual: %s", value)
		t.Fail()
	}
}

func TestStore(t *testing.T) {
	lsm := Connect("files/2026-11-09.txt")
	defer lsm.Close()
	if ok := lsm.Store("key3", "value3"); !ok {
		t.Fail()
	}
	expected := "value3"
	actual := lsm.Get("key3")
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}
