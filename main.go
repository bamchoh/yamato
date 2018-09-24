package main

import (
	"os"
	"path/filepath"
)

func main() {
	z := NewZipper()
	path := `C:\Users\bamch\go\src`
	z.Execute(path)
	f, err := os.Create(filepath.Base(path) + ".zip")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(z.Bytes())
}
