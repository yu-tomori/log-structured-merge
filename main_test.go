package main

import (
	"os"
	"path/filepath"
	"testing"
)

// don't flush, just keep in memory
func connect(t *testing.T, content string) *LogStructuredMergeTree {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.log")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	lsm, err := Connect(dir)
	if err != nil {
		t.Fatal(err)
	}

	return lsm
}

func TestGet(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	value, err := lsm.Get("key2")
	if err != nil {
		t.Errorf("lsm.Get failed: %s", value)
		t.Fail()
	}
	if value != "value2" {
		t.Errorf("expected: value2, actual: %s", value)
		t.Fail()
	}
}

func TestGetDuplicateKeys(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
key1,value
`
	lsm := connect(t, content)
	actual, err := lsm.Get("key1")
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
		t.Fail()
	}
	// get latest value
	expected := "value"
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
		t.Fail()
	}
}

func TestStore(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	if ok := lsm.Store("key3", "value3"); !ok {
		t.Fail()
	}
	expected := "value3"
	actual, err := lsm.Get("key3")
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
		t.Fail()
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStoreAndGet(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	if ok := lsm.Store("key4", "value5"); !ok {
		t.Fail()
	}
	value, err := lsm.Get("key4")
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
		t.Fail()
	}
	expected := "value5"
	if value != expected {
		t.Errorf("expected: %s, actual: %s", expected, value)
		t.Fail()
	}
}
