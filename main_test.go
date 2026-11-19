package main

import (
	"os"
	"path/filepath"
	"testing"
)

// don't flush, just keep in memory
func connect(t *testing.T, content string) LogStructuredMergeTree {
	t.Helper()
	path := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return Connect(path)
}

func TestGet(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	defer lsm.file.Close()
	value := lsm.Get("key2")
	if value != "value2" {
		t.Errorf("expected: value2, actual: %s", value)
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

func TestStoreAndGet(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	defer lsm.file.Close()
	if ok := lsm.Store("key4", "value5"); !ok {
		t.Fail()
	}
	value := lsm.Get("key4")
	expected := "value5"
	if value != expected {
		t.Errorf("expected: %s, actual: %s", expected, value)
		t.Fail()
	}
}

func TestGetDuplicateKeys(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
key1,value3
`
	lsm := connect(t, content)
	defer lsm.file.Close()
	value := lsm.Get("key1")
	// get latest value
	if value != "value3" {
		t.Errorf("expected: value3, actual: %s", value)
		t.Fail()
	}
}
