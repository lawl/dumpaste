package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const BLOBPATH = "data/blobs/"
const EXPIRYPATH = "data/expiries/"
const HTTPSTOREPATH = "/store"
const HTTPGETPATH = "/get/"
const RESOURCEPATH = "/c/"
const MAINPATH = "/"
const UIFILE = "../client/index.html"
const UIPATH = "../client/"
const UMASK = 0664

const MAX_STORAGE_HOURS = 64

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
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "500 Internal Server oopsie", http.StatusInternalServerError)
		return
	}
	id := getID(content)

	var expHrs int64
	expString := "-1"
	expStrings, ok := r.URL.Query()["exp"]
	if ok && len(expStrings[0]) > 0 {
		expString = expStrings[0]
	}
	expHrs, err = strconv.ParseInt(expString, 10, 64)
	if err != nil {
		expHrs = -1
	}
	if expHrs > MAX_STORAGE_HOURS {
		expHrs = MAX_STORAGE_HOURS
	}
	// requesting never expires but file is > 100kb, store for max duration instead
	if expHrs < 1 && len(content) > 1024*100 {
		expHrs = MAX_STORAGE_HOURS
	}
	if expHrs > 0 {
		ioutil.WriteFile(EXPIRYPATH+strconv.Itoa(int(timestamp()+(60*60*expHrs))), []byte(id), UMASK)
	}
	ioutil.WriteFile(BLOBPATH+id, content, UMASK)
	fmt.Fprintf(w, "OK:%s", id)
}

func timestamp() int64 {
	return time.Now().Unix()
}

func gethandler(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[len(HTTPGETPATH):]
	cleanName := filepath.FromSlash(fname)
	c, err := ioutil.ReadFile(BLOBPATH + cleanName)
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

func garbageCollector() {
	for {
		time.Sleep(time.Minute)
		fmt.Println("Collecting garbage")
		files, err := filepath.Glob(EXPIRYPATH + "*")
		if err != nil {
			continue
		}

		for _, f := range files {
			_, fname := filepath.Split(f)
			stamp, err := strconv.Atoi(fname)
			if err != nil {
				continue
			}
			if time.Now().Unix() > int64(stamp) && int64(stamp) > 0 {
				fmt.Println("deleting file", f)
				garbageFile, _ := ioutil.ReadFile(EXPIRYPATH + fname)
				os.Remove(EXPIRYPATH + fname)
				os.Remove(BLOBPATH + string(garbageFile))
			}
		}
	}
}

func main() {
	go garbageCollector()
	port := os.Args[1]
	http.HandleFunc(HTTPSTOREPATH, storehandler)
	http.HandleFunc(HTTPGETPATH, gethandler)
	http.HandleFunc(RESOURCEPATH, resourcehandler)
	http.HandleFunc(MAINPATH, mainhandler)
	http.ListenAndServe(port, nil)
}
