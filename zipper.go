package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"runtime"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
)

type Zipper struct {
}

func (z *Zipper) Execute(path string, w io.Writer) (err error) {
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		os.Chdir(cd)
		runtime.GC()
	}()
	os.Chdir(path)

	files, err := zglob.Glob(`**\*`)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, s := range files {
		if err = z.addToZip(s, zipWriter); err != nil {
			return err
		}
	}

	return nil
	// eg := errgroup.Group{}
	// for _, s := range files {
	// 	eg.Go(func() error {
	// 		defer func() {
	// 			if err := recover(); err != nil {
	// 				log.Println("[Zipper Goroutine]", err)
	// 			}
	// 		}()
	// 		return z.addToZip(s, zipWriter)
	// 	})
	// }

	// if err := eg.Wait(); err != nil {
	// 	return err
	// }
	// return nil
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
