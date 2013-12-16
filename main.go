package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
)

type fileInfo struct {
	Filename string // file name
	Uploaded string // uploaded time
	Size     string // size
	Filetype string // fill type: f=file; d=dir;
}

type fileInfos []fileInfo

func (this fileInfos) Len() int      { return len(this) }
func (this fileInfos) Swap(i, j int) { this[i], this[j] = this[j], this[i] }
func (this fileInfos) Less(i, j int) bool {
	if this[i].Filetype == this[j].Filetype {
		return strings.ToLower(this[i].Filename) < strings.ToLower(this[j].Filename)
	}
	return this[i].Filetype < this[j].Filetype
}

type FileServerHandler struct {
	root string
}

func (this FileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	root := http.Dir(this.root)
	path := path.Clean("/" + r.URL.Path)
	log.Printf("%s > GET %s", r.RemoteAddr, path)

	f, err := root.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	d, _ := f.Stat()
	if d.IsDir() {
		list := listFile(f)
		// log.Printf("%v", rows)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t := template.New("index")
		t, err := t.ParseFiles("index.tmpl")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = t.ExecuteTemplate(w, "index.tmpl", list)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

func listFile(f http.File) []fileInfo {
	var list []fileInfo
	for {
		dirs, err := f.Readdir(100)
		if err != nil || len(dirs) == 0 {
			break
		}
		for _, d := range dirs {
			name := d.Name()
			time := d.ModTime().Format("2006-01-02")
			size := formatSize(d.Size())
			filetype := "f"
			if d.IsDir() {
				name += "/"
				size = "-"
				filetype = "d"
			}
			list = append(list, fileInfo{name, time, size, filetype})
		}
	}
	sort.Sort(fileInfos(list))
	return list
}

func formatSize(size int64) string {
	unit := []int64{1024, 1024 * 1024, 1024 * 1024 * 1024}
	if size >= unit[2] {
		i := size / unit[2]
		return strconv.FormatInt(i, 10) + " GB"
	} else if size >= unit[1] {
		i := size / unit[1]
		return strconv.FormatInt(i, 10) + " MB"
	} else if size >= unit[0] {
		i := size / unit[0]
		return strconv.FormatInt(i, 10) + " KB"
	}
	return strconv.FormatInt(size, 10) + " B"
}

var (
	server string
	fsPath string
)

func main() {
	log.Fatal(http.ListenAndServe(server, FileServerHandler{fsPath}))
}

func init() {
	flag.StringVar(&server, "l", ":8080", "http listen port")
	flag.StringVar(&fsPath, "p", "C:/tmp", "server path")
	flag.Parse()
	fmt.Printf("listen %s, fs %s\r\n", server, fsPath)
}
