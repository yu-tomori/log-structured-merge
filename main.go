package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Storage interface {
	Get(key string) (value string)
	Store(key string, value string) (ok bool, savedAt int64)
}

type LogStructuredMergeTree struct {
	dir   string
	files []struct { // new file added at last
		path  string           // open, close in every access
		index map[string]int64 // key to offset in file
	}
}

const maxBytesPerFile = 10000

func Connect(connectURI string) (*LogStructuredMergeTree, error) {
	if err := os.MkdirAll(connectURI, 0755); err != nil {
		return nil, err
	}

	lsm := &LogStructuredMergeTree{
		dir: connectURI,
		files: make([]struct {
			path  string
			index map[string]int64 // key to offset in file
		}, 0, 1),
	}

	entries, err := os.ReadDir(connectURI)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".log" {
			continue
		}
		path := filepath.Join(connectURI, entry.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		if err := lsm.indexFile(f); err != nil {
			return nil, fmt.Errorf("failed to index a file: %v", err)
		}
		f.Close()
	}

	return lsm, nil
}

func (lsm *LogStructuredMergeTree) indexFile(f *os.File) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset seek: %v\n", err)
	}

	abs, err := filepath.Abs(f.Name())
	if err != nil {
		return fmt.Errorf("failed to get absolute path of file: %v", err)
	}

	file := struct {
		path  string
		index map[string]int64
	}{
		path:  abs,
		index: make(map[string]int64),
	}
	lsm.files = append(lsm.files, file)

	// build index
	reader := bufio.NewReader(f)
	var offset int64
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			record := strings.TrimSuffix(line, "\n")
			record = strings.TrimSuffix(line, "\r")
			fields := strings.SplitN(record, ",", 2)
			k := fields[0]
			file.index[k] = int64(offset)
			offset += int64(len(line))
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("unexpected error while building index: %v", err)
		}
	}
}

// return empty string, if key not found
func (lsm *LogStructuredMergeTree) Get(key string) (value string, err error) {
	for _, file := range slices.Backward(lsm.files) {
		f, err := os.Open(file.path)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to reset seek: %v\n", err)
		}

		info, err := f.Stat()
		if err != nil {
			return "", fmt.Errorf("failed to stat fiel: %v\n", err)
		}

		offset, ok := file.index[key]
		start := int64(offset)
		if !ok {
			continue
		}
		r := bufio.NewReader(io.NewSectionReader(
			f,
			start,
			info.Size()-start,
		))

		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read line: %v\n", err)
		}
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		fields := strings.SplitN(line, ",", 2)
		if len(fields) != 2 || fields[0] != key {
			return "", nil
		}
		return fields[1], nil
	}
	// no key found
	return "", nil
}

func (lsm *LogStructuredMergeTree) Store(key string, value string) error {
	latest := lsm.files[len(lsm.files)-1]
	f, err := os.OpenFile(latest.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}()

	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	// for index updates
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get filesize: %v", err)
	}
	offset := info.Size()

	line := fmt.Sprintf("%s,%s", key, value)
	if int(offset)+len([]byte(line)) > maxBytesPerFile {
		fileName := fmt.Sprintf("%s.log", time.Now().Format(time.RFC3339))
		file, err := os.Create(lsm.dir + fileName)
		if err != nil {
			return err
		}

		// write new line to new file
		if _, err := fmt.Fprintln(file, line); err != nil {
			return fmt.Errorf("failed to flush: %v\n", err)
		}

		if err := lsm.indexFile(file); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(f, line); err != nil {
		return fmt.Errorf("failed to flush: %v\n", err)
	}
	latest.index[key] = int64(offset)

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("invalid arguements: length %d", len(os.Args))
		os.Exit(1)
	}

	connectPath := os.Args[1]
	lsm, err := Connect(connectPath)
	if err != nil {
		fmt.Printf("failed to connect URI: %v", err)
		os.Exit(1)
	}

	value, err := lsm.Get(os.Args[2])
	if err != nil {
		fmt.Printf("failed to get value: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", os.Args[2], value)
}
