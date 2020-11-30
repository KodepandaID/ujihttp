package benchmark

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/KodepandaID/ujihttp"
	"github.com/KodepandaID/ujihttp/internal/fasthttp"
	"github.com/KodepandaID/ujihttp/pkg/histogram"
)

var body = &bytes.Buffer{}
var writer *multipart.Writer

var (
	timeout   int64
	reqError  int64
	size      int64
	start     time.Time
	totalReq  int64
	respOK    int64
	respNotOK int64
)

// ReqBench is a request benchmark config
type ReqBench struct {
	method       string
	path         string
	headers      ujihttp.H
	cookies      ujihttp.H
	send         ujihttp.H
	sendJSONData ujihttp.JSON
	sendFile     bool
	contentType  string
	debug        bool
	concurrent   int
	duration     int
	pipeline     int
	timeout      int
}

// New to start a benchmark test
func New() *ReqBench {
	return &ReqBench{
		concurrent: 10,
		duration:   10,
		pipeline:   1,
		timeout:    10,
	}
}

// Concurrent to set total concurrent connections to use
func (rb *ReqBench) Concurrent(c int) *ReqBench {
	rb.concurrent = c

	return rb
}

// Duration to set the number of seconds to run the benchmark
func (rb *ReqBench) Duration(d int) *ReqBench {
	rb.duration = d

	return rb
}

// Pipeline to set the number of pipelined requests to use
func (rb *ReqBench) Pipeline(p int) *ReqBench {
	rb.pipeline = p

	return rb
}

// Timeout to set the number of timout requests
func (rb *ReqBench) Timeout(t int) *ReqBench {
	rb.timeout = t

	return rb
}

// GET request method
func (rb *ReqBench) GET(p string) *ReqBench {
	rb.method = "GET"
	rb.path = p

	return rb
}

// POST request method
func (rb *ReqBench) POST(p string) *ReqBench {
	rb.method = "POST"
	rb.path = p

	return rb
}

// PUT request method
func (rb *ReqBench) PUT(p string) *ReqBench {
	rb.method = "PUT"
	rb.path = p

	return rb
}

// DELETE request method
func (rb *ReqBench) DELETE(p string) *ReqBench {
	rb.method = "DELETE"
	rb.path = p

	return rb
}

// PATCH request method
func (rb *ReqBench) PATCH(p string) *ReqBench {
	rb.method = "PATCH"
	rb.path = p

	return rb
}

// HEAD request method
func (rb *ReqBench) HEAD(p string) *ReqBench {
	rb.method = "HEAD"
	rb.path = p

	return rb
}

// OPTIONS request method
func (rb *ReqBench) OPTIONS(p string) *ReqBench {
	rb.method = "OPTIONS"
	rb.path = p

	return rb
}

// WithHeader to set header request
func (rb *ReqBench) WithHeader(h ujihttp.H) *ReqBench {
	rb.headers = h

	return rb
}

// WithCookies to set cookies request
func (rb *ReqBench) WithCookies(h ujihttp.H) *ReqBench {
	rb.cookies = h

	return rb
}

// WithContentType to set content-type request
func (rb *ReqBench) WithContentType(ct string) *ReqBench {
	rb.contentType = ct

	return rb
}

// SendJSON to send json data
func (rb *ReqBench) SendJSON(j ujihttp.JSON) *ReqBench {
	rb.sendJSONData = j
	rb.contentType = "application/json"

	return rb
}

// SendFormData to send multipart/form-data
func (rb *ReqBench) SendFormData(h ujihttp.H) *ReqBench {
	rb.send = h

	return rb
}

// SendFile (fieldName, filepath string)
//
// to send file from filepath
func (rb *ReqBench) SendFile(fn, path string) *ReqBench {
	rb.sendFile = true

	if body.Len() == 0 {
		writer = multipart.NewWriter(body)
	}

	f, e := os.Open(path)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	form, e := writer.CreateFormFile(fn, filepath.Base(path))
	if e != nil {
		panic(e)
	}
	io.Copy(form, f)

	rb.contentType = writer.FormDataContentType()

	return rb
}

// SendMultipleFile (fieldName, filepath string)
//
// Use [] on the fieldname to send multiple file from filepath
func (rb *ReqBench) SendMultipleFile(fn string, path []string) *ReqBench {
	rb.sendFile = true

	if body.Len() == 0 {
		writer = multipart.NewWriter(body)
	}

	for _, p := range path {
		f, e := os.Open(p)
		if e != nil {
			panic(e)
		}
		defer f.Close()

		form, e := writer.CreateFormFile(fn, filepath.Base(p))
		if e != nil {
			panic(e)
		}
		io.Copy(form, f)
	}

	rb.contentType = writer.FormDataContentType()

	return rb
}

// Run to start a benchmark test
func (rb *ReqBench) Run() {
	fmt.Printf("Running %ds test @ %s\n", rb.duration, rb.path)
	if rb.pipeline == 1 {
		fmt.Printf("%d connections\n\n", rb.concurrent)
	} else {
		fmt.Printf("%d connections with %d pipelining factor\n\n", rb.concurrent, rb.pipeline)
	}

	callHTTP(rb)

	fmt.Println("Req/Bytes counts sampled once per second.")
	fmt.Printf("%s requests in %.2fs, %s read\n", countRequest(totalReq), math.Round(time.Since(start).Seconds()), countReadRequest(size))
	fmt.Printf("%s 2xx responses and %s non 2xx responses\n", countRequest(respOK), countRequest(reqError))
	fmt.Printf("%s errors (%s timeouts)\n", countRequest(reqError), countRequest(timeout))
}

func callHTTP(rb *ReqBench) {
	latency := histogram.New()
	reqBytes := histogram.New().SetTimeSleep(rb.duration)

	u, _ := url.Parse(rb.path)
	start = time.Now()
	for i := 0; i < rb.concurrent; i++ {
		c := fasthttp.PipelineClient{
			Addr:               fmt.Sprintf("%v:%v", u.Hostname(), u.Port()),
			MaxPendingRequests: rb.pipeline,
			IsTLS:              u.Scheme == "https",
		}
		for j := 0; j < rb.pipeline; j++ {
			go func() {
				req := fasthttp.AcquireRequest()
				defer fasthttp.ReleaseRequest(req)
				req.Header.SetMethod(rb.method)
				req.SetRequestURI(rb.path)

				if len(rb.headers) > 0 {
					for key, val := range rb.headers {
						req.Header.Set(key, val)
					}
				}

				if len(rb.cookies) > 0 {
					for key, val := range rb.cookies {
						req.Header.SetCookie(key, val)
					}
				}

				if rb.contentType != "" {
					req.Header.Set("Content-Type", rb.contentType)
				}
				req.Header.Set("User-Agent", "UjiHTTP/Benchmark")

				resp := fasthttp.AcquireResponse()
				defer fasthttp.ReleaseResponse(resp)
				for {
					totalReq++
					if e := c.DoTimeout(req, resp, time.Second*time.Duration(rb.timeout)); e != nil {
						latency.AddTime(time.Since(start))
						reqError++
						if e == fasthttp.ErrTimeout {
							timeout++
						}
					} else {
						latency.AddTime(time.Since(start))
						size += int64(len(resp.Body()))
						resp.Header.VisitAll(func(key, value []byte) {
							size += int64(len(key) + len(value))
						})
						reqBytes.AddSize(time.Since(start), size)

						if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
							respOK++
						} else {
							reqError++
						}
						resp.Reset()
					}
				}
			}()
		}
	}
	time.Sleep(time.Duration(rb.duration) * time.Second)
	latency.CalcLatency()
	reqBytes.CalcReqBytes()
}

func countRequest(c int64) string {
	var total string
	if c < 1000 {
		total = fmt.Sprintf("%d", c/1000)
	} else if c >= 1000 {
		total = fmt.Sprintf("%dk", c/1000)
	}

	return total
}

func countReadRequest(s int64) string {
	var size string
	if s >= 1000000 {
		size = fmt.Sprintf("%.1f MB", float32(s/1000000))
	} else if s >= 1000 && s < 1000000 {
		size = fmt.Sprintf("%.1f KB", float32(s/1000))
	} else {
		size = fmt.Sprintf("%d B", s)
	}

	return size
}
