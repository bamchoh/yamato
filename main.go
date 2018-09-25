package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	template = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>yamato</title>
</head>
<body>
%s
</body>
</html>
`
	basePath = `C:\Users\bamch\go`
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func defaultHandler(w http.ResponseWriter, r *http.Request, path string) {
	cd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer os.Chdir(cd)
	os.Chdir(path)
	paths, err := filepath.Glob("*")
	if err != nil {
		w.Write([]byte("error!!"))
		return
	}

	body := ""
	for _, p := range paths {
		body += fmt.Sprintf(`<a href='/?path=%s'>%s</a><br />`, p, p)
	}

	w.Write([]byte(fmt.Sprintf(template, body)))
}

func zipHandler(w http.ResponseWriter, r *http.Request, path string) {
	zipFilename := filepath.Base(path) + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+zipFilename)
	z := NewZipper()
	z.Execute(path)
	w.Write(z.Bytes())
	runtime.GC()
}

func handler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()["path"]) > 0 {
		dir := r.URL.Query()["path"][0]
		zipHandler(w, r, filepath.Join(basePath, dir))
	} else {
		defaultHandler(w, r, basePath)
	}
}
