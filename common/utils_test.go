package common

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGin(t *testing.T) {
	app := GetGinApp()
	app.GET("/api/sample", WrapHandler(func(c *gin.Context) (interface{}, error) {
		return 1, nil
	}))
	app.GET("/api/error", WrapHandler(func(c *gin.Context) (interface{}, error) {
		return nil, errors.New("err1")
	}))
	getResultString := func(api string, body io.Reader) string {
		req, _ := http.NewRequest("GET", api, body)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		return string(w.Body.Bytes())
	}
	assert.Equal(t, "{\"msg\":\"pong\"}", getResultString("/api/ping", nil))
	assert.Equal(t, "1", getResultString("/api/sample", nil))
	assert.Equal(t, "{\"code\":500,\"message\":\"err1\"}", getResultString("/api/error", strings.NewReader("{}")))
}

func TestFuncs(t *testing.T) {
	wd := MustGetwd()
	assert.NotEqual(t, "", wd)

	dir1 := GetCallerCodeDir()
	assert.Equal(t, true, strings.HasSuffix(dir1, "common"))

}
