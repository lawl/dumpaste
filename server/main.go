package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const DATAPATH = "data/"
const HTTPSTOREPATH = "/store"
const HTTPGETPATH = "/get/"
const RESOURCEPATH = "/c/"
const MAINPATH = "/"
const UIFILE = "../client/index.html"
const UIPATH = "../client/"
const UMASK = 0664

func getID(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	md := hash.Sum(nil)
	mdStr := base64.URLEncoding.EncodeToString(md)
	return mdStr
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func storehandler(w http.ResponseWriter, r *http.Request) {
	//content := []byte(r.FormValue("r"))
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "500 Internal Server oopsie", http.StatusInternalServerError)
		return
	}
	id := getID(content)
	ioutil.WriteFile(DATAPATH+id, content, UMASK)
	fmt.Fprintf(w, "OK:%s", id)
}
func gethandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(HTTPGETPATH):]
	cleanName := filepath.FromSlash(fname)
	c, err := ioutil.ReadFile(DATAPATH + cleanName)
	if err == nil {
		fmt.Fprint(w, string(c))
		return
	}
	fmt.Fprint(w, "FAIL")
}

func mainhandler(w http.ResponseWriter, r *http.Request) {
	ui, err := ioutil.ReadFile(UIFILE)
	if err != nil {
		fmt.Fprint(w, "Couldn't read user interface")
		return
	}
	fmt.Fprint(w, string(ui))
}

func resourcehandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(RESOURCEPATH):]
	if fname == "" {
		http.Error(w, "404 File not found", http.StatusNotFound)
		return
	}
	if strings.Contains(fname, "..") || fname[:1] == "/" {
		http.Error(w, "blyat", http.StatusForbidden)
		return
	}
	fname = UIPATH + fname
	if fileExists(fname) {
		if strings.HasSuffix(fname, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		}
		if strings.HasSuffix(fname, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		ui, err := ioutil.ReadFile(fname)
		if err == nil {
			fmt.Fprint(w, string(ui))
			return
		}
	}
	http.Error(w, "404 File not found", http.StatusNotFound)
}

func main() {
	port := os.Args[1]
	http.HandleFunc(HTTPSTOREPATH, storehandler)
	http.HandleFunc(HTTPGETPATH, gethandler)
	http.HandleFunc(RESOURCEPATH, resourcehandler)
	http.HandleFunc(MAINPATH, mainhandler)
	http.ListenAndServe(port, nil)
}
