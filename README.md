
curl request:

~~~
curl -XPOST -T image.jpeg -H 'Content-Type: multipart/form-data' http://192.168.33.11:8080/split
~~~

HTTP POST form:

~~~
http://192.168.33.11:8080/
~~~

TODO:  

1. http://www.alexedwards.net/blog/serving-static-sites-with-go - serve like this
2. Use tokens: https://github.com/dgrijalva/jwt-go
3. see below or use http.ServeMux

~~~
package main

import (
    "fmt"
    "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "HomeHandler")
}

func serveSingle(pattern string, filename string) {
    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filename)
    })
}

func main() {
    http.HandleFunc("/", HomeHandler) // homepage

    // Mandatory root-based resources
    serveSingle("/sitemap.xml", "./sitemap.xml")
    serveSingle("/favicon.ico", "./favicon.ico")
    serveSingle("/robots.txt", "./robots.txt")

    // Normal resources
    http.Handle("/static", http.FileServer(http.Dir("./static/")))

    http.ListenAndServe(":8080", nil)
}
