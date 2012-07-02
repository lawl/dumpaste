package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const COUNTERFILE = "data/count"
const DATAPATH = "data/"
const HTTPSTOREPATH = "/store"
const HTTPGETPATH = "/get/"
const RESOURCEPATH = "/c/"
const MAINPATH = "/"
const UIFILE = "../client/index.html"
const UIPATH = "../client/"
const UMASK = 0664

func GetId() int {
	id := 0
	c, err := ioutil.ReadFile(COUNTERFILE)
	if err == nil {
		id, err = strconv.Atoi(string(c))
		if err != nil {
			id = 0
		}
	}
	id += 1
	ioutil.WriteFile(COUNTERFILE, []byte(strconv.Itoa(id)), UMASK)
	return id
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func Storehandler(w http.ResponseWriter, r *http.Request) {
	id := GetId()
	ioutil.WriteFile(DATAPATH+strconv.Itoa(id), []byte(r.FormValue("r")), UMASK)
	fmt.Fprintf(w, "OK:%d", id)
}
func Gethandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(HTTPGETPATH):]
	if _, err := strconv.Atoi(fname); err == nil {
		c, err := ioutil.ReadFile(DATAPATH + fname)
		if err == nil {
			fmt.Fprint(w, c)
			return
		}
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
    if fname== "" {
		http.Error(w, "404 File not found", http.StatusNotFound)
        return;
    }
	if strings.Contains(fname, "..") || fname[:1] == "/" {
		http.Error(w, "Nice try. Now fuck off", http.StatusForbidden)
		return
	}
    fname = UIPATH + fname;
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
	http.HandleFunc(HTTPSTOREPATH, Storehandler)
	http.HandleFunc(HTTPGETPATH, Gethandler)
	http.HandleFunc(RESOURCEPATH, Resourcehandler)
	http.HandleFunc(MAINPATH, Mainhandler)
	http.ListenAndServe(":8080", nil)
}
