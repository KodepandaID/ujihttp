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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Login binding
type Login struct {
	User     string                `form:"user" json:"user"`
	Password string                `form:"password" json:"password"`
	File     *multipart.FileHeader `form:"file"`
}

// ginEngine to used gin router
func ginEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!!")
	})

	r.GET("/delay", func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "Hello World!!")
		c.Done()
	})

	r.POST("/post-json", func(c *gin.Context) {
		var payload Login
		c.ShouldBindJSON(&payload)

		c.JSON(http.StatusOK, gin.H{
			"user":     payload.User,
			"password": payload.Password,
		})
		c.Done()
	})

	r.POST("/post-form", func(c *gin.Context) {
		var payload Login
		c.ShouldBind(&payload)

		c.JSON(http.StatusOK, gin.H{
			"user":     payload.User,
			"password": payload.Password,
		})
		c.Done()
	})

	r.POST("/post-file", func(c *gin.Context) {
		file, e := c.FormFile("file")
		if e != nil {
			fmt.Println(e)
			c.Status(http.StatusBadRequest)
			c.Done()
		}

		if file != nil {
			c.String(http.StatusOK, "Success")
			c.Done()
		}
	})

	r.POST("/post-form-with-file", func(c *gin.Context) {
		var payload Login
		c.ShouldBind(&payload)

		file, e := c.FormFile("file")
		if e != nil {
			c.Status(http.StatusBadRequest)
			c.Done()
		}

		if file != nil {
			c.JSON(http.StatusOK, gin.H{
				"user":     payload.User,
				"password": payload.Password,
			})
		}
	})

	r.POST("/post-multiple-file", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["file[]"]

		if len(files) == 2 {
			c.String(http.StatusOK, "Success")
			c.Done()
		} else {
			c.Status(http.StatusBadRequest)
			c.Done()
		}
	})

	return r
}

func TestGinGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/").
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestGinWithDelayGET(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		GET("/delay").
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestGinJsonPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-json").
		SendJSON(ujihttp.JSON{
			"user":     "test",
			"password": "password",
		}).
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestGinFormPOST(t *testing.T) {
	r := ujihttp.New()

	r.
		SetDebug(true).
		POST("/post-form").
		SendFormData(ujihttp.H{
			"user":     "test",
			"password": "password",
		}).
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestGinUploadFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-file").
		SendFile("file", path+"/assets/sample.jpg").
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}

func TestGinFormWithUploadFile(t *testing.T) {
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
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			resp := Login{}
			json.Unmarshal([]byte(rec.Body.String()), &resp)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "test", resp.User)
			assert.Equal(t, "password", resp.Password)
		})
}

func TestGinUploadMultipleFile(t *testing.T) {
	r := ujihttp.New()

	path, _ := os.Getwd()

	r.
		SetDebug(true).
		POST("/post-multiple-file").
		SendMultipleFile("file[]", []string{path + "/assets/sample.jpg", path + "/assets/sample.jpg"}).
		Run(ginEngine(), func(req *http.Request, rec *httptest.ResponseRecorder) {
			assert.Equal(t, http.StatusOK, rec.Code)
		})
}
