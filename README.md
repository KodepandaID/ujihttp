# ujiHTTP
HTTP handler testing for golang web framework. The benchmark features used [fasthttp](https://github.com/valyala/fasthttp).

## Support Framework
Will support another framework for the future
* [x] [Http Handler](https://golang.org/pkg/net/http/)
* [x] [Gin](https://github.com/gin-gonic/gin)
* [x] [Echo](https://github.com/labstack/echo)
* [x] [Mux](https://github.com/gorilla/mux)

## Installation
```bash
go get github.com/KodepandaID/ujihttp
```

## Example
This is an example of the main program using Gin, but you can also use net/http, Gorilla Mux, or Echo.

### main.go
```go
package main

import (
    "net/http"

    "github.com/KodepandaID/ujihttp"
    "github.com/gin-gonic/gin"
)

func GinEngine() *gin.Engine {
	r := gin.New()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!!")
    })

    return r
}
```

This is an example for testing the main program.

### test.go
```go
package main

import (
    "net/http"
    "testing"

    "github.com/KodepandaID/ujihttp"
    "github.com/stretchr/testify/assert"
)

func TestGinGET(t *testing.T) {
    r := ujihttp.New()

	r.
		GET("/").
		Run(GinEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
        })
}
```

## Usage

### `SetDebug`
To enable debug mode.

```go
func main() {
    r := ujihttp.New()

    r.
        SetDebug(true).
        GET("/").
        Run(GinEngine())
}
```

### `METHOD`
You can use the GET, POST, PUT, DELETE, PATCH, and HEAD method.

```go
func main() {
    r := ujihttp.New()

	r.
        GET("/").
        Run(GinEngine())
}
```

### `SendJSON`
If you use the POST or PUT method, you can send a JSON body.

```go
func main() {
    r := ujihttp.New()

	r.
        POST("/add-json").
        SendJSON(ujihttp.JSON{
			"user":     "test",
			"password": "password",
		}).
        Run(GinEngine())
}
```

### `SendFormData`
If you want to send data with header `multipart/form-data`

```go
func main() {
    r := ujihttp.New()

	r.
        POST("/add-json").
        SendJSON(ujihttp.JSON{
			"user":     "test",
			"password": "password",
		}).
        Run(GinEngine())
}
```

### `SendFile`
You can combine this method with SendFormData method.

```go
func main() {
    r := ujihttp.New()
	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-file").
        SendFile("file", path+"/assets/sample.jpg").
        Run(GinEngine())
}
```

### `SendMultipleFile`
If you want to upload multiple files, use this method. You can combine this method with SendFormData method.

```go
func main() {
    r := ujihttp.New()
	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-file").
        SendMultipleFile("file", path+"/assets/sample.jpg").
        Run(GinEngine())
}
```

## How to Benchmark
You can benchmark your API using this library. But, if you want to benchmark your API, make sure your API already run.

```go
func main() {
    r := benchmark.New()
	r.
		Duration(10).
		Concurrent(500).
		Pipeline(1).
		GET("http://localhost:3000").
		Run()
}
```

See below the methods you can use on the Benchmark features.

### `Duration`
You can set how long the benchmark run in the second.
```go
r := benchmark.New()
r.Duration(10)
```

### `Concurrent`
You can set how many the concurrent connection.
```go
r := benchmark.New()
r.Concurrent(500)
```

#### `Pipeline`
You can set how many used the HTTP pipeline request.
```go
r := benchmark.New()
r.Pipeline(500)
```

#### `Methods`
Same with the API testing handler, you can use GET, POST, PUT, DELETE, PATCH, and HEAD method.
```go
r := benchmark.New()
r.GET("/").Run()
```

## License
Copyright [Yudha Pratama](https://github.com/lordaur). Licensed under [MIT](./LICENSE).
