package main

import (
	// "bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const TEMP = "_TMP"

func main() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH env var not set.")
	}
	path := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "Emyrk", "LendingBot", "messages")
	tmpPath := filepath.Join(path, "tmp")

	err := os.RemoveAll(tmpPath)
	if err != nil {
		panic(fmt.Sprintf("Error removing dir: %s", err.Error()))
	}
	err = os.Mkdir(tmpPath, os.ModePerm)
	if err != nil {
		panic("Error making tmp dir.")
	}
	defer os.Remove(tmpPath)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(fmt.Sprintf("Could not read dir[%s]", path))
	}

	baseFile, err := NewBaseFile(filepath.Join(path, "index.en"))
	if err != nil {
		panic(fmt.Sprintf("Error reading in base file: %s", err.Error()))
	}

	for _, f := range files {
		if f.Name() == "index.en" || !strings.Contains(f.Name(), "index") {
			continue
		}

		nf, err := NewFile(filepath.Join(path, f.Name()))
		if err != nil {
			panic(fmt.Sprintf("Error reading in file[%s], error: %s", f.Name(), err.Error()))
		}

		err = nf.ProcessLines(baseFile.Lines)
		if err != nil {
			panic(fmt.Sprintf("Error processing lines file[%s], error: %s", f.Name(), err.Error()))
		}

		tmpFile, err := os.Create(filepath.Join(tmpPath, f.Name()))
		if err != nil {
			panic(fmt.Sprintf("Error creating tmp file[%s], error: %s", f.Name(), err.Error()))
		}

		_, err = tmpFile.WriteString(nf.EmptyLines())
		if err != nil {
			panic(fmt.Sprintf("Error writing string for file[%s], error: %s", f.Name(), err.Error()))
		}
	}

	for _, f := range files {
		if f.Name() == "index.en" || !strings.Contains(f.Name(), "index") {
			continue
		}
		err := os.Rename(filepath.Join(tmpPath, f.Name()), filepath.Join(path, f.Name()))
		if err != nil {
			fmt.Printf("Error moving file[%s], note other files have been moved.\n", f.Name())
		} else {
			fmt.Printf("Success moving file[%s]\n", f.Name())
		}
	}
}

type BaseFile struct {
	Lines []string
}

func NewBaseFile(fn string) (*BaseFile, error) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	f := &BaseFile{
		Lines: strings.Split(string(b), "\n"),
	}
	return f, nil
}

type File struct {
	extraLines    map[string]string //key (left side), value (right side)
	originalLines []string          //the original lines
	tmpLines      []string          //the current tmp lines
}

func NewFile(originalFn string) (*File, error) {
	b1, err := ioutil.ReadFile(originalFn)
	if err != nil {
		return nil, err
	}
	return &File{
		extraLines:    make(map[string]string),
		originalLines: strings.Split(string(b1), "\n"),
		tmpLines:      make([]string, 0),
	}, nil
}

func (f *File) ProcessLines(newLines []string) error {
	var (
		i int64
		b bool
	)
	for _, l := range newLines {
		if len(l) == 0 {
			f.tmpLines = append(f.tmpLines, l)
			continue
		} else if l[:1] == "#" {
			f.tmpLines = append(f.tmpLines, l)
			continue
		}
		sArr := strings.SplitN(l, "=", 2)
		if len(sArr) != 2 {
			return fmt.Errorf("Error with split: %s", sArr)
		}
		if i, b = f.processKey(sArr[0], i); !b {
			f.tmpLines = append(f.tmpLines, l)
		}
	}

	for a := i; a < int64(len(f.originalLines)); a++ {
		f.extraLines[f.originalLines[a]] = ""
	}
	return nil
}

func (f *File) processKey(key string, index int64) (int64, bool) {
	if v, ok := f.extraLines[key]; ok {
		f.tmpLines = append(f.tmpLines, strings.Trim(key, " ")+"="+v)
		return index, true
	}

	for i := index; i < int64(len(f.originalLines)); i++ {
		kArr := strings.SplitN(f.originalLines[i], "=", 2)
		if len(kArr) != 2 {
			continue
		} else if kArr[0] != key {
			f.extraLines[kArr[0]] = kArr[1]
			continue
		}
		f.tmpLines = append(f.tmpLines, f.originalLines[i])
		i++
		return i, true
	}
	return index, false
}

func (f *File) EmptyLines() string {
	fs := ""
	for _, s := range f.tmpLines {
		fs += s + "\n"
	}
	// for k, v := range f.extraLines {
	// 	if v != "" {
	// 		fs += k + "=" + v + "\n"
	// 	} else {
	// 		fs += k + "\n"
	// 	}
	// }

	return fs
}
