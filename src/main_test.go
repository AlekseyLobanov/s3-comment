package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testPreview(t *testing.T, app *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/preview",
		strings.NewReader("{\"text\":\"Hello, *dear* __world__\"}"),
	)
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(
		t,
		"{\"text\":\"<p>Hello, <em>dear</em> <strong>world</strong></p>\"}",
		strings.TrimSpace(strings.ReplaceAll(w.Body.String(), "\\n", "")),
	)
}

func testCount(t *testing.T, app *gin.Engine) {
	for _, method := range []string{"OPTIONS", "POST"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(
			method,
			"/count",
			nil,
		)
		app.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		if method == "POST" {
			assert.Equal(t, "[]", strings.TrimSpace(w.Body.String()))
		}
	}

}

func TestEngineE2e(t *testing.T) {
	app := GetGinApp()

	t.Run("TestWebPreview", func(t *testing.T) {
		testPreview(t, app)
	})

	t.Run("TestWebCount", func(t *testing.T) {
		testCount(t, app)
	})
}
