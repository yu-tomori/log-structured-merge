package main

import (
	"os"
	"io"
	"strings"
	"time"
	"fmt"
	"bufio"
)

type byteOffset int64

type Storage interface {
	Get(key string) (value string)
	Store(key string, value string) (ok bool, savedAt byteOffset)
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
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
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

func (lsm LogStructuredMergeTree) Close() {
	lsm.file.Close()
}

// return empty string, if key not found
func (lsm LogStructuredMergeTree) Get(key string) (value string) {
	if _, err := lsm.file.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("failed to reset seek: %v\n", err)
		return ""
	}

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

func (lsm LogStructuredMergeTree) Store(key string, value string) (ok bool) {
	line := fmt.Sprintf("%s,%s", key, value)
	if _, err := fmt.Fprintln(lsm.file, line); err != nil {
		fmt.Printf("failed to store: %v\n", err)
		return false
	}
	return true
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("invalid arguements: length %d", len(os.Args))
		os.Exit(1)
	}

	connectPath := os.Args[1]
	lsm := Connect(fmt.Sprintf("files/%s", connectPath))
	defer lsm.Close()

	fmt.Printf("%s: %s\n", os.Args[2], lsm.Get(os.Args[2]))
}

