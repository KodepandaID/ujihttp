package ujihttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/KodepandaID/ujihttp/pkg/cli"
)

const version = "1.0.0"

var body = &bytes.Buffer{}
var writer *multipart.Writer

// H is a HTTP map string data
type H map[string]string

// JSON is HTTP map string interface data
type JSON map[string]interface{}

// ReqConf is a request config to test api
type ReqConf struct {
	method       string
	path         string
	headers      H
	cookies      H
	send         H
	sendJSONData JSON
	sendFile     bool
	contentType  string
	debug        bool
	assertStatus int
}

// ResponseFunc response handling func type
type ResponseFunc func(*http.Request, *httptest.ResponseRecorder)

// New to start api test handler
func New() *ReqConf {
	return &ReqConf{}
}

// SetDebug to enable debug mode
func (rc *ReqConf) SetDebug(b bool) *ReqConf {
	rc.debug = b

	return rc
}

// GET request method
func (rc *ReqConf) GET(p string) *ReqConf {
	rc.method = "GET"
	rc.path = p

	return rc
}

// POST request method
func (rc *ReqConf) POST(p string) *ReqConf {
	rc.method = "POST"
	rc.path = p

	return rc
}

// PUT request method
func (rc *ReqConf) PUT(p string) *ReqConf {
	rc.method = "PUT"
	rc.path = p

	return rc
}

// DELETE request method
func (rc *ReqConf) DELETE(p string) *ReqConf {
	rc.method = "DELETE"
	rc.path = p

	return rc
}

// PATCH request method
func (rc *ReqConf) PATCH(p string) *ReqConf {
	rc.method = "PATCH"
	rc.path = p

	return rc
}

// HEAD request method
func (rc *ReqConf) HEAD(p string) *ReqConf {
	rc.method = "HEAD"
	rc.path = p

	return rc
}

// OPTIONS request method
func (rc *ReqConf) OPTIONS(p string) *ReqConf {
	rc.method = "OPTIONS"
	rc.path = p

	return rc
}

// WithHeader to set header request
func (rc *ReqConf) WithHeader(h H) *ReqConf {
	rc.headers = h

	return rc
}

// WithCookies to set cookies request
func (rc *ReqConf) WithCookies(h H) *ReqConf {
	rc.cookies = h

	return rc
}

// WithContentType to set content-type request
func (rc *ReqConf) WithContentType(ct string) *ReqConf {
	rc.contentType = ct

	return rc
}

// SendJSON to send json data
func (rc *ReqConf) SendJSON(j JSON) *ReqConf {
	rc.sendJSONData = j
	rc.contentType = "application/json"

	return rc
}

// SendFormData to send multipart/form-data
func (rc *ReqConf) SendFormData(h H) *ReqConf {
	rc.send = h

	return rc
}

// SendFile (fieldName, filepath string)
//
// to send file from filepath
func (rc *ReqConf) SendFile(fn, path string) *ReqConf {
	rc.sendFile = true

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

	rc.contentType = writer.FormDataContentType()

	return rc
}

// SendMultipleFile (fieldName, filepath string)
//
// Use [] on the fieldname to send multiple file from filepath
func (rc *ReqConf) SendMultipleFile(fn string, path []string) *ReqConf {
	rc.sendFile = true

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

	rc.contentType = writer.FormDataContentType()

	return rc
}

// Run to start api test
func (rc *ReqConf) Run(r http.Handler, response ResponseFunc) {
	if len(rc.send) > 0 {
		if body.Len() == 0 {
			writer = multipart.NewWriter(body)
		}

		for key, val := range rc.send {
			writer.WriteField(key, val)
		}

		rc.contentType = writer.FormDataContentType()
	}

	if len(rc.sendJSONData) > 0 {
		js, e := json.Marshal(rc.sendJSONData)
		if e != nil {
			panic(e)
		}
		body = bytes.NewBuffer(js)
	}

	if body.Len() > 0 && len(rc.sendJSONData) == 0 {
		writer.Close()
	}

	req, _ := http.NewRequest(rc.method, rc.path, body)

	if len(rc.headers) > 0 {
		for key, val := range rc.headers {
			req.Header.Set(key, val)
		}
	}

	if len(rc.cookies) > 0 {
		for key, val := range rc.cookies {
			req.AddCookie(&http.Cookie{
				Name:  key,
				Value: val,
			})
		}
	}

	if rc.contentType != "" {
		req.Header.Set("Content-Type", rc.contentType)
	}
	req.Header.Set("User-Agent", "UjiHTTP/"+version)

	rec := httptest.NewRecorder()
	startTime := time.Now()
	r.ServeHTTP(rec, req)
	endTime := time.Now().Sub(startTime)

	response(req, rec)

	if rc.debug {
		b, _ := ioutil.ReadAll(rec.Body)
		bs := len(b)
		cli.WriteDebug(&cli.DebugData{
			Method:     rc.method,
			Path:       rc.path,
			Duration:   endTime,
			BodySize:   bs,
			Code:       rec.Code,
			CodeStatus: http.StatusText(rec.Code),
		})
	}
}
