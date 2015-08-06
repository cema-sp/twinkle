package main

import (
	// "fmt"
	"net/http"
	"log"
	"io/ioutil"
)

// type PostHandler struct {}

// func (h PostHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

func handlePost(rw http.ResponseWriter, req *http.Request) {
	var err error
	if req.Method != "POST" {
		log.Println("Wrong method: ", req.Method)
		return
	}

	err = http.ErrNotMultipart
	for _, v := range req.Header["Content-Type"] {
		if v == "multipart/form-data" {
			err = nil
		}
	}
	if err != nil {
		log.Println(err)
		return
	}

	if req.ContentLength <= 0 {
		log.Println("Body length: 0")
        return
	}

	var body []byte
	body, err = ioutil.ReadAll(req.Body)
    if err != nil {
        log.Println(err)
        return
    }

    if len(body) == 0 {
    	log.Println("Body length: 0")
        return
    }

	log.Println(string(body))
}

func main() {
    http.HandleFunc("/split", handlePost)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
