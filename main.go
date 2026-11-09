package main

import (
	"os"
	"strings"
	"time"
	"fmt"
	"bufio"
)

type byteOffset int64

type Storage interface {
	Get(key string) (value string)
	Store(key string) (ok bool, savedAt byteOffset)
}

type LogStructuredMergeTree struct {
	file *os.File
	index map[string]byteOffset
}

func Init() (*os.File) {
	t := time.Now().Format("2006-01-02")
	filePath := fmt.Sprintf("files/%s", t)
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("failed to initialize file: %v", err)
		os.Exit(1)
	}
	return f
}

func Connect(path string) (LogStructuredMergeTree) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("not found path: %s", path)
	}
	i := make(map[string]byteOffset)
	lsm := LogStructuredMergeTree{
		file: f,
		index: i,
	}
	return lsm
}

// return empty string, if key not found
func (lsm LogStructuredMergeTree) Get(key string) (value string) {
	// full scan now. optimize it later.
	scanner := bufio.NewScanner(lsm.file)
	kvs := make(map[string]string)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		if len(line) != 2 {
			fmt.Printf("malformed key value: %v", line)
			os.Exit(1)
		}
		kvs[line[0]] = line[1]
	}
	value, ok := kvs[key]
	if !ok {
		return ""
	}
	return value
}

func (lsm LogStructuredMergeTree) Store(key string) (ok bool, savedAt byteOffset) {
	return true, byteOffset(int64(0))
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("invalid arguements: length %d", len(os.Args))
		os.Exit(1)
	}

	connectPath := os.Args[1]
	lsm := Connect(fmt.Sprintf("files/%s", connectPath))

	fmt.Printf("%s: %s\n", os.Args[2], lsm.Get(os.Args[2]))
}

