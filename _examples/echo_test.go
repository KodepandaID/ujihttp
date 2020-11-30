package examples

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KodepandaID/ujihttp"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

// Login binding
type Login struct {
	User     string                `form:"user" json:"user"`
	Password string                `form:"password" json:"password"`
	File     *multipart.FileHeader `form:"file"`
}

func echoEngine() *echo.Echo {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/delay", func(c echo.Context) error {
		time.Sleep(2 * time.Second)
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/post-json", func(c echo.Context) error {
		req := new(Login)
		c.Bind(req)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"user":     req.User,
			"password": req.Password,
		})
	})

	e.POST("/post-form", func(c echo.Context) error {
		req := new(Login)
		c.Bind(req)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"user":     req.User,
			"password": req.Password,
		})
	})

	e.POST("/post-file", func(c echo.Context) error {
		file, e := c.FormFile("file")
		if e != nil {
			fmt.Println(e)
			return c.NoContent(http.StatusBadRequest)
		}

		if file == nil {
			return c.NoContent(http.StatusBadRequest)
		}

		return c.String(http.StatusOK, "Success")
	})

	e.POST("/post-form-with-file", func(c echo.Context) error {
		req := new(Login)
		c.Bind(req)

		file, e := c.FormFile("file")
		if e != nil {
			fmt.Println(e)
			return c.NoContent(http.StatusBadRequest)
		}

		if file == nil {
			return c.NoContent(http.StatusBadRequest)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"user":     req.User,
			"password": req.Password,
		})
	})

	e.POST("/post-multiple-file", func(c echo.Context) error {
		form, _ := c.MultipartForm()
		files := form.File["file[]"]

		if len(files) == 2 {
			return c.String(http.StatusOK, "Success")
		}

		return c.NoContent(http.StatusBadRequest)
	})

	return e
}

func TestEchoGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/").
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestEchoWithDelayGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/delay").
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestEchoJsonPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-json").
		SendJSON(ujihttp.JSON{
			"user":     "test",
			"password": "password",
		}).
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestEchoFormPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-form").
		SendFormData(ujihttp.H{
			"user":     "test",
			"password": "password",
		}).
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestEchoUploadFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-file").
		SendFile("file", path+"/assets/sample.jpg").
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestEchoFormWithUploadFile(t *testing.T) {
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
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestEchoUploadMultipleFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-multiple-file").
		SendMultipleFile("file[]", []string{path + "/assets/sample.jpg", path + "/assets/sample.jpg"}).
		Run(echoEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}
