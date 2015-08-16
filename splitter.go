package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gographics/imagick/imagick"
	"html/template"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	// "github.com/cema-sp/twinkle"
	// "runtime"
)

// type PostHandler struct {}
// func (h PostHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

// TODO: Configuration in ENV
const (
	maxImageSize       = (4 * 2) << 20    // 4 Mb
	htmlFormFileField  = "imageFile"      // image field name
	htmlFormTokenField = "token"          // token field name
	maxImageChunks     = 16               // chunks amount
	minChunksPerSide   = 4                // for chunks radius calculation
	chunkersCnt        = 1                // # of workers
	tokenSize          = 4                // length of token
	tokenExpiration    = 5 * time.Minute  // token validity time
	tokenDestroy       = 10 * time.Minute // token live time
	tokenCleanup       = 5 * time.Minute  // token store cleanup frequency
)

var (
	tokenStore map[string]time.Time
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

	token := req.FormValue(htmlFormTokenField)
	err = validateToken(token)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Token is ok")
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

func scheduleChunking(imagesChan chan []byte) error {
	for imageBytes := range imagesChan {
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
			x, err := rand.Int(rand.Reader, big.NewInt(int64(image.GetImageWidth()-d)))
			if err != nil {
				log.Println(err)
				continue
			}
			y, err := rand.Int(rand.Reader, big.NewInt(int64(image.GetImageHeight()-d)))
			if err != nil {
				log.Println(err)
				continue
			}

			chunkers.Add(1)
			go func() {
				defer chunkers.Done()

				if err := createChunk(mask.Clone(), image, int(x.Int64()), int(y.Int64())); err != nil {
					log.Println(fmt.Sprintf("Chunking for (%d;%d) failed: %v", x, y, err))
				}
			}()
		}
		chunkers.Wait()
	}

	return nil
}

func serveStaticFile(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func generateToken() (t string, err error) {
	randBytes := make([]byte, tokenSize)

	_, err = rand.Read(randBytes)
	if err != nil {
		return
	}

	t = base64.URLEncoding.EncodeToString(randBytes)
	if _, ok := tokenStore[t]; ok {
		t, err = generateToken()
		return
	}

	tokenStore[t] = time.Now().Add(tokenExpiration)
	return
}

func validateToken(token string) error {
	expires, ok := tokenStore[token]
	if !ok {
		err := ErrTokenNotFound(token)
		return &err
	}

	delete(tokenStore, token)

	if time.Now().After(expires) {
		err := ErrTokenExpired(token)
		return &err
	}

	return nil
}

func tokenStoreCleaner(c <-chan time.Time) {
	for _ = range c {
		for key, exp := range tokenStore {
			if time.Now().After(exp) {
				delete(tokenStore, key)
			}
		}
	}
}

func init() {
	// TODO: Stronger Rand
	// rand.Seed(time.Now().UTC().UnixNano())
	// Use all CPUs (TODO: Profile it)
	// runtime.GOMAXPROCS(runtime.NumCPU())

	tokenStore = make(map[string]time.Time)
}

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	// Start chunking workers
	imagesChan := make(chan []byte)
	for i := 0; i < chunkersCnt; i++ {
		go scheduleChunking(imagesChan)
	}

	// Start token store cleaner
	go tokenStoreCleaner(time.Tick(tokenCleanup))

	// Serve Assets
	http.Handle("/css/", http.FileServer(http.Dir("./assets")))
	http.Handle("/js/", http.FileServer(http.Dir("./assets")))
	http.Handle("/fonts/", http.FileServer(http.Dir("./assets")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlePost(imagesChan, w, r)
		case "GET":
			// TODO: function
			token, err := generateToken()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), 500)
			}

			tmpl, err := template.ParseFiles("index.html")
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), 500)
			}
			tmpl.Execute(w, token)
		default:
			http.NotFound(w, r)
		}
		return
	})

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type ErrTokenNotFound string

func (e ErrTokenNotFound) Error() string {
	return fmt.Sprintf("Token '%s' not found", string(e))
}

type ErrTokenExpired string

func (e ErrTokenExpired) Error() string {
	return fmt.Sprintf("Token '%s' expired", string(e))
}
