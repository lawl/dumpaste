package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const COUNTERFILE = "data/count"
const DATAPATH = "data/"
const HTTPSTOREPATH = "/store"
const HTTPGETPATH = "/get/"
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
			fmt.Fprintf(w, "%s", c)
			return
		}
	}
	fmt.Fprintf(w, "%s", "FAIL")
}

func main() {
	http.HandleFunc(HTTPSTOREPATH, Storehandler)
	http.HandleFunc(HTTPGETPATH, Gethandler)
	http.ListenAndServe(":8080", nil)
}
