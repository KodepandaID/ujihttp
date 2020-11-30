package examples

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KodepandaID/ujihttp"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Login binding
type Login struct {
	User     string                `form:"user" json:"user"`
	Password string                `form:"password" json:"password"`
	File     *multipart.FileHeader `form:"file"`
}

func muxEngine() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}).Methods("GET")

	r.HandleFunc("/delay", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}).Methods("GET")

	r.HandleFunc("/post-json", func(w http.ResponseWriter, rr *http.Request) {
		b, _ := ioutil.ReadAll(rr.Body)
		req := Login{}
		json.Unmarshal(b, &req)

		json.NewEncoder(w).Encode(map[string]string{
			"user":     req.User,
			"password": req.Password,
		})
	}).Methods("POST")

	r.HandleFunc("/post-form", func(w http.ResponseWriter, rr *http.Request) {
		user := rr.FormValue("user")
		password := rr.FormValue("password")

		json.NewEncoder(w).Encode(map[string]string{
			"user":     user,
			"password": password,
		})
	}).Methods("POST")

	r.HandleFunc("/post-file", func(w http.ResponseWriter, rr *http.Request) {
		_, file, e := rr.FormFile("file")
		if e != nil {
			fmt.Println(e)
			w.WriteHeader(http.StatusBadRequest)
		}

		if file != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		}
	}).Methods("POST")

	r.HandleFunc("/post-form-with-file", func(w http.ResponseWriter, rr *http.Request) {
		user := rr.FormValue("user")
		password := rr.FormValue("password")

		_, file, e := rr.FormFile("file")
		if e != nil {
			fmt.Println(e)
			w.WriteHeader(http.StatusBadRequest)
		}

		if file != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"user":     user,
				"password": password,
			})
		}
	}).Methods("POST")

	r.HandleFunc("/post-multiple-file", func(w http.ResponseWriter, rr *http.Request) {
		rr.ParseMultipartForm(100)
		form := rr.MultipartForm
		files := form.File["file[]"]

		if len(files) == 2 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}).Methods("POST")

	return r
}

func TestMuxGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/").
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestMuxWithDelayGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/delay").
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestMuxJsonPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-json").
		SendJSON(ujihttp.JSON{
			"user":     "test",
			"password": "password",
		}).
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestMuxFormPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-form").
		SendFormData(ujihttp.H{
			"user":     "test",
			"password": "password",
		}).
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestMuxUploadFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-file").
		SendFile("file", path+"/assets/sample.jpg").
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestMuxFormWithUploadFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-form-with-file").
		SendFormData(ujihttp.H{
			"user":     "test",
			"password": "password",
		}).
		SendFile("file", path+"/assets/sample.jpg").
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestMuxUploadMultipleFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-multiple-file").
		SendMultipleFile("file[]", []string{path + "/assets/sample.jpg", path + "/assets/sample.jpg"}).
		Run(muxEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}
