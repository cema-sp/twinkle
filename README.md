
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
2. Token Store Watcher (time.Ticker)
3. input accept & server validation
