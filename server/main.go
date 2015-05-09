package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
    "crypto/sha256"
    "encoding/base64"
    "path/filepath"
)

const DATAPATH = "data/"
const HTTPSTOREPATH = "/store"
const HTTPGETPATH = "/get/"
const RESOURCEPATH = "/c/"
const MAINPATH = "/"
const UIFILE = "../client/index.html"
const UIPATH = "../client/"
const UMASK = 0664

func GetId(data []byte) string {
    hash := sha256.New()
    hash.Write(data)
    md := hash.Sum(nil)
    mdStr := base64.URLEncoding.EncodeToString(md)
	return mdStr
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func Storehandler(w http.ResponseWriter, r *http.Request) {
    content := []byte(r.FormValue("r"));
	id := GetId(content)
	ioutil.WriteFile(DATAPATH+id, content, UMASK)
	fmt.Fprintf(w, "OK:%s", id)
}
func Gethandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(HTTPGETPATH):]
    cleanName := filepath.FromSlash(fname);
	c, err := ioutil.ReadFile(DATAPATH + cleanName)
	if err == nil {
		fmt.Fprint(w, string(c))
		return
	}
	fmt.Fprint(w, "FAIL")
}

func Mainhandler(w http.ResponseWriter, r *http.Request) {
	ui, err := ioutil.ReadFile(UIFILE)
	if err != nil {
		fmt.Fprint(w, "Couldn't read user interface")
		return
	}
	fmt.Fprint(w, string(ui))
}

func Resourcehandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(RESOURCEPATH):]
	if fname == "" {
		http.Error(w, "404 File not found", http.StatusNotFound)
		return
	}
	if strings.Contains(fname, "..") || fname[:1] == "/" {
		http.Error(w, "Nice try. Now fuck off", http.StatusForbidden)
		return
	}
	fname = UIPATH + fname
	if FileExists(fname) {
		ui, err := ioutil.ReadFile(fname)
		if err == nil {
			fmt.Fprint(w, string(ui))
			return
		}
	}
	http.Error(w, "404 File not found", http.StatusNotFound)
}

func main() {
    port := os.Args[1];
	http.HandleFunc(HTTPSTOREPATH, Storehandler)
	http.HandleFunc(HTTPGETPATH, Gethandler)
	http.HandleFunc(RESOURCEPATH, Resourcehandler)
	http.HandleFunc(MAINPATH, Mainhandler)
	http.ListenAndServe(port, nil)
}
