package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
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

func (z *Zipper) Execute(path string) (err error) {
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

	eg, ctx := errgroup.WithContext(context.TODO())
	for _, s := range files {
		eg.Go(func() error {
			return z.addToZip(s, zipWriter, ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (z *Zipper) addToZip(filename string, zipWriter *zip.Writer, ctx context.Context) error {
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

	errCh := make(chan error, 1)
	go func() {
		z.m.Lock()
		defer z.m.Unlock()

		writer, err := zipWriter.Create(filename)
		if err != nil {
			errCh <- errors.Wrap(err, "[CREATE]")
		}

		_, err = io.Copy(writer, src)
		if err != nil {
			errCh <- errors.Wrap(err, fmt.Sprintf("[COPY](%s)", src.Name()))
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
