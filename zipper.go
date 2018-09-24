package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
)

func NewZipper() *Zipper {
	var buf []byte
	m := new(sync.Mutex)
	b := bytes.NewBuffer(buf)
	return &Zipper{m, b}
}

type Zipper struct {
	m *sync.Mutex
	b *bytes.Buffer
}

func (z *Zipper) Bytes() []byte {
	return z.b.Bytes()
}

func (z *Zipper) Execute(path string) {
	cd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer os.Chdir(cd)
	os.Chdir(path)
	files, err := zglob.Glob(`**\*`)
	if err != nil {
		panic(err)
	}

	zipWriter := zip.NewWriter(z.b)
	defer zipWriter.Close()

	wg := &sync.WaitGroup{}
	for _, s := range files {
		go func(filename string) {
			wg.Add(1)
			var err1 error
			if err1 = z.addToZip(filename, zipWriter); err1 != nil {
				panic(err1)
			}
			wg.Done()
		}(s)
	}
	wg.Wait()
}

func (z *Zipper) addToZip(filename string, zipWriter *zip.Writer) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return errors.Wrap(err, "isDir : os.Stat : ")
	}

	if fi.Mode().IsDir() {
		return nil
	}

	src, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer src.Close()

	z.m.Lock()

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return errors.Wrap(err, "[CREATE]")
	}

	_, err = io.Copy(writer, src)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[COPY](%s)", src.Name()))
	}

	z.m.Unlock()

	return nil
}
