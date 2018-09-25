package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func NewZipper(w io.Writer) *Zipper {
	m := new(sync.Mutex)
	return &Zipper{m, w}
}

type Zipper struct {
	m *sync.Mutex
	b io.Writer
}

func (z *Zipper) Execute(path string, ctx context.Context) (err error) {
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cd)
	os.Chdir(path)
	files, err := zglob.Glob(`**\*`)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(z.b)
	defer zipWriter.Close()

	errCh := make(chan error, len(files))
	defer close(errCh)

	eg := errgroup.Group{}
	for _, s := range files {
		eg.Go(func() error {
			child, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				errCh <- z.addToZip(s, zipWriter)
			}()

			select {
			case <-child.Done():
				return child.Err()
			case err := <-errCh:
				return err
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.Println(err)
		return err
	}
	return nil
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
	defer z.m.Unlock()

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return errors.Wrap(err, "[CREATE]")
	}

	_, err = io.Copy(writer, src)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[COPY](%s)", src.Name()))
	}

	return nil
}
