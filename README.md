
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
2. input accept & server validation
3. Style file input: http://plugins.krajee.com/file-input OR
http://www.abeautifulsite.net/whipping-file-inputs-into-shape-with-bootstrap-3/
4. Handle GET with regexp (for id)
5. Create data model (for mongo)
6. Implement data model (in go types)
7. Configure ORM
8. Save image to mongo (full)
9. Load image from mongo & display for user
