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

func TestEP(t *testing.T) {
	skipped := true
	err := func() (rerr error) {
		defer P2E(&rerr)
		E2P(errors.New("err1"))
		skipped = false
		return nil
	}()
	assert.Equal(t, true, skipped)
	assert.Equal(t, "err1", err.Error())
	err = CatchP(func() {
		PanicIf(true, errors.New("err2"))
	})
	assert.Equal(t, "err2", err.Error())
	err = func() (rerr error) {
		defer func() {
			x := recover()
			assert.Equal(t, 1, x)
		}()
		defer P2E(&rerr)
		panic(1)
	}()
}

func TestTernary(t *testing.T) {
	assert.Equal(t, "1", OrString("", "", "1"))
	assert.Equal(t, "", OrString("", "", ""))
	assert.Equal(t, "1", If(true, "1", "2"))
	assert.Equal(t, "2", If(false, "1", "2"))
}

func TestMarshal(t *testing.T) {
	a := 0
	type e struct {
		A int
	}
	e1 := e{A: 10}
	m := map[string]int{}
	assert.Equal(t, "1", MustMarshalString(1))
	assert.Equal(t, []byte("1"), MustMarshal(1))
	MustUnmarshal([]byte("2"), &a)
	assert.Equal(t, 2, a)
	MustUnmarshalString("3", &a)
	assert.Equal(t, 3, a)
	MustRemarshal(&e1, &m)
	assert.Equal(t, 10, m["A"])
}

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

func TestResty(t *testing.T) {
	resp, err := RestyClient.R().Get("http://baidu.com")
	assert.Equal(t, nil, err)
	err2 := CatchP(func() {
		CheckRestySuccess(resp, err)
	})
	assert.NotEqual(t, nil, err2)
}
