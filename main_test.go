package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestConnectEmptyDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	lsm, err := Connect(dir)
	if err != nil {
		t.Fatal(err)
	}
	key := "key1"
	value, err := lsm.Get(key)
	if err != nil {
		t.Fatal(value)
	}
	if value != "" {
		t.Errorf("value should be empty string, but %s", value)
	}

	expected := "value1"
	if err := lsm.Store(key, expected); err != nil {
		t.Fatal(err)
	}
	actual, err := lsm.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
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

func TestCompact(t *testing.T) {
	file1 := `key1,value1
key1,value1
key1,value1
key1,value1
key2,value1
key2,value2
key2,value3
key2,value4
key2,value5
key3,value3
`
	file2 := `key3,value4
key4,value8
`

	file3 := `key5,value10
key6,value11
`

	lsm := connect(t, file1, file2, file3)
	if err := lsm.Compact(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(lsm.files[0].path)
	if err != nil {
		t.Fatal(err.Error())
	}
	actual := make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		fields := strings.SplitN(line, ",", 2)
		if len(fields) != 2 {
			t.Fatal(err)
		}
		actual[fields[0]] = fields[1]
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"key1": "value1",
		"key2": "value5",
		"key3": "value4",
		"key4": "value8",
	}
	if len(actual) != len(expected) {
		t.Errorf("actual length: %d, expected length: %d", len(actual), len(expected))
	}
	for k, v := range expected {
		if actual[k] != v {
			t.Errorf("%s != %s", actual[k], v)
		}
	}

	cases := []struct {
		key      string
		expected string
	}{
		{"key1", "value1"},
		{"key2", "value5"},
		{"key3", "value4"},
		{"key4", "value8"},
		{"key5", "value10"},
		{"key6", "value11"},
	}
	for _, c := range cases {
		actual, err := lsm.Get(c.key)
		if err != nil {
			t.Errorf("failed to get %s: %v", c.key, err)
			continue
		}
		if actual != c.expected {
			t.Errorf("%s != %s", actual, expected)
		}
	}

}
