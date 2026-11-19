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
	i := buildIndex(f)
	lsm := LogStructuredMergeTree{
		file: f,
		index: i,
	}
	return lsm
}

func buildIndex(f *os.File) (index map[string]byteOffset) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		fmt.Printf("failed to reset seek: %v\n", err)
		os.Exit(1)
	}
	index = make(map[string]byteOffset)

	reader := bufio.NewReader(f)
	var offset int64
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			record := strings.TrimSuffix(line, "\n")
			record = strings.TrimSuffix(line, "\r")
			fields := strings.SplitN(record, ",", 2)
			k := fields[0]
			index[k] = byteOffset(offset)
			offset += int64(len(line))
		}
		if err != nil {
			if err == io.EOF {
				return index
			}
			fmt.Printf("unexpected error while building index: %v", err)
			return index
		}
	}
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

	info, err := lsm.file.Stat()
	if err != nil {
		fmt.Printf("failed to stat fiel: %v\n", err)
		return ""
	}

	offset, ok := lsm.index[key]
	start := int64(offset)
	if !ok {
		return ""
	}
	r := bufio.NewReader(io.NewSectionReader(
		lsm.file,
		start,
		info.Size()-start,
	))

	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("failed to read line: %v\n", err)
		return ""
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	fields := strings.SplitN(line, ",", 2)
	if len(fields) != 2 || fields[0] != key {
		return ""
	}

	return fields[1]
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

