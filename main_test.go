package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// don't flush, just keep in memory
func connect(t *testing.T, contents ...string) *LogStructuredMergeTree {
	t.Helper()

	dir := t.TempDir()
	for n, c := range contents {
		fileName := fmt.Sprintf("fixture_%d.log", n+1)
		path := filepath.Join(dir, fileName)
		if err := os.WriteFile(path, []byte(c), 0600); err != nil {
			t.Fatal(err)
		}
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
		t.Fatalf("lsm.Get failed: %s", value)
	}
	if value != "value2" {
		t.Errorf("expected: value2, actual: %s", value)
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
		t.Fatalf("failed to lsm.Get: %v", err)
	}
	// get latest value
	expected := "value"
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestGetMultipleFile(t *testing.T) {
	file1 := `
key1,value1
key2,value2
key3,value3
`
	file2 := `
key2,value5
key3,value6
`
	lsm := connect(t, file1, file2)

	key := "key3"
	expected := "value6"
	actual, err := lsm.Get(key)
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStore(t *testing.T) {
	content := `
key1,value1
key2,value2
key3,value3
`
	lsm := connect(t, content)
	if err := lsm.Store("key3", "value3"); err != nil {
		t.FailNow()
	}
	expected := "value3"
	actual, err := lsm.Get("key3")
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStoreMultipleFile(t *testing.T) {
	file1 := `
key1,value1
key2,value2
key3,value3
`
	file2 := `
key2,value5
key3,value6
`
	lsm := connect(t, file1, file2)

	key := "key3"
	expected := "value4"
	if err := lsm.Store(key, expected); err != nil {
		t.FailNow()
	}
	actual, err := lsm.Get(key)
	if err != nil {
		t.Errorf("failed to lsm.Get: %v", err)
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestStoreBeyondMaxBytes(t *testing.T) {
	content := ""
	n := 0
	for {
		line := fmt.Sprintf("key%d,value%d\n", n, n)
		if len(content)+len(line) > maxBytesPerFile {
			break
		}
		content = content + line
		n++
	}
	t.Logf("maxBytes: %d, fixture size: %d\n", maxBytesPerFile, len(content))
	if len(content) > maxBytesPerFile {
		t.Fatal("precondition isn't sutisfied\n")
	}

	lsm := connect(t, content)
	key := "key"
	expected := "valuevaluevaluevaluevaluevaluevaluevalue"
	if err := lsm.Store(key, expected); err != nil {
		t.Fatalf("failed to store: %v", err)
	}
	actual, err := lsm.Get(key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}
