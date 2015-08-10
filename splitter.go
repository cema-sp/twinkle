package main

import (
	"errors"
	"fmt"
	"github.com/gographics/imagick/imagick"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	// "runtime"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	// "html/template"
)

// type PostHandler struct {}

// func (h PostHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

// TODO: Configuration in ENV
const (
	maxImageSize      = 8 << 20 // 4 Mb
	htmlFormFileField = "imageFile"
	maxImageChunks    = 16
	minChunksPerSide  = 4
)

func matchContentType(ct []string, matching string) error {
	ctJoined := strings.Join(ct, "; ")
	matched, err := regexp.MatchString(matching+`;*\s*.*`, ctJoined)
	if !matched {
		if err != nil {
			log.Println(err)
		}
		return errors.New(
			fmt.Sprintf(
				"Invalid Content-Type: %s, Expected: %s",
				ctJoined,
				matching))
	}
	return nil
}

func handlePost(imagesChan chan []byte, rw http.ResponseWriter, req *http.Request) {
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

	// size <= maxImageSize - store in MEM
	if err = req.ParseMultipartForm(maxImageSize); err != nil {
		log.Println(err)
		return
	}

	file, fileHandler, err := req.FormFile(htmlFormFileField)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// Validate file type
	if err = matchContentType(
		fileHandler.Header["Content-Type"],
		`image/jpeg`); err != nil {
		log.Println(err)
		return
	}

	// log.Println(fileHandler.Header)

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

	imagesChan <- image

	// if err = scheduleChunking(image); err != nil {
	// 	log.Println(err)
	// 	return
	// }

	log.Println("Chunking scheduled: ", len(image))
}

func calcDnR(image *imagick.MagickWand) (d uint, r float64) {
	width := image.GetImageWidth()
	height := image.GetImageHeight()

	if height > width {
		d = width / minChunksPerSide
	} else {
		d = height / minChunksPerSide
	}
	r = float64(d) / 2
	return
}

func createMask(d uint, r float64) (mask *imagick.MagickWand) {
	mask = imagick.NewMagickWand()

	pw := imagick.NewPixelWand()
	defer pw.Destroy()
	dw := imagick.NewDrawingWand()
	defer dw.Destroy()

	pw.SetColor("none")
	mask.NewImage(d, d, pw)
	pw.SetColor("white")
	dw.SetFillColor(pw)
	dw.Circle(r, r, r-1, r*2-1)
	mask.DrawImage(dw)

	return
}

func createChunk(mask, image *imagick.MagickWand, x, y int) error {
	if err := mask.CompositeImage(image, imagick.COMPOSITE_OP_SRC_IN, -x, -y); err != nil {
		return err
		// fmt.Sprintf("Chunking for (%d;%d) failed: %v", x, y, err)
	}

	if err := mask.WriteImage("tmp/" + strconv.Itoa(x) + "_" + strconv.Itoa(y) + ".png"); err != nil {
		return err
		// fmt.Sprintf("Saving chunk for (%d;%d) failed: %v", x, y, err)
	}
	return nil
}

func scheduleChunking(imageBytes []byte) error {
	// outputFileName := "tmp/tmp.jpeg"

	image := imagick.NewMagickWand()
	defer image.Destroy()

	if err := image.ReadImageBlob(imageBytes); err != nil {
		return err
	}

	d, r := calcDnR(image)
	mask := createMask(d, r)
	defer mask.Destroy()

	// TODO: Profile it, mb faster in consequence
	var chunkers sync.WaitGroup
	for i := 0; i < maxImageChunks; i++ {
		x := rand.Intn(int(image.GetImageWidth() - d))
		y := rand.Intn(int(image.GetImageHeight() - d))

		chunkers.Add(1)
		go func() {
			defer chunkers.Done()

			if err := createChunk(mask.Clone(), image, x, y); err != nil {
				log.Println(fmt.Sprintf("Chunking for (%d;%d) failed: %v", x, y, err))
			}
		}()
	}
	chunkers.Wait()
	return nil
}

func init() {
	// TODO: Stronger Rand
	rand.Seed(time.Now().UTC().UnixNano())
	// Use all CPUs (TODO: Profile it)
	// runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	imagesChan := make(chan []byte)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "form.html")
	})
	http.HandleFunc("/split", func(w http.ResponseWriter, r *http.Request) {
		handlePost(imagesChan, w, r)
	})
	// http.HandleFunc("/split", handlePost)

	log.Fatal(http.ListenAndServe(":8080", nil))

	log.Println("Started")
}
