package main

import (
	"errors"
	"fmt"
	// "github.com/gographics/imagick/imagick"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	// "html/template"
)

// type PostHandler struct {}

// func (h PostHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

func matchContentType(ct []string, matching string) error {
	ctJoined := strings.Join(ct, "; ")
	matched, err := regexp.MatchString(matching+`;*\s*.*`, ctJoined)
	if !matched {
		if err != nil {
			log.Println(err)
		}
		return errors.New(
			fmt.Sprintf(
				"Invalid Content-Type: %s,\nExpected: %s",
				ctJoined,
				matching))
	}
	return nil
}

func handlePost(rw http.ResponseWriter, req *http.Request) {
	var err error
	if req.Method != "POST" {
		log.Println("Wrong method: ", req.Method)
		return
	}

	if err = matchContentType(
		req.Header["Content-Type"],
		`multipart/form-data`); err != nil {
		log.Println(err)
		return
	}

	if req.ContentLength <= 0 {
		log.Println("Body length: 0")
		return
	}

	file, fileHeader, err := req.FormFile("imageFile")
	if err != nil {
		log.Println(err)
		return
	}

	// Validate file type
	if err = matchContentType(
		fileHeader.Header["Content-Type"],
		`image/jpeg`); err != nil {
		log.Println(err)
		return
	}

	var image []byte
	image, err = ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}

	// Validate true image type
	if err = matchContentType(
		[]string{http.DetectContentType(image)},
		`image/jpeg`); err != nil {
		log.Println(err)
		return
	}

	// if len(image) == 0 {
	// 	log.Println("Body length: 0")
	//   return
	// }

	log.Println("Image: ", len(image), http.DetectContentType(image))
}

func main() {
	// imagick.Initialize()
	// defer imagick.Terminate()

	// mw := imagick.NewMagickWand()
	// defer mw.Destroy()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "form.html")
	})
	http.HandleFunc("/split", handlePost)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
