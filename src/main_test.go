package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewGeneration(t *testing.T) {
	app := GetGinApp()

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
