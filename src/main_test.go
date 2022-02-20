package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func s(val string) *string {
	return &val
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

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

func TestEngineWithoutIntegrations(t *testing.T) {
	app := GetGinApp(ApplicationConfig{})

	t.Run("TestWebPreview", func(t *testing.T) {
		testPreview(t, app)
	})

	t.Run("TestWebCount", func(t *testing.T) {
		testCount(t, app)
	})
}

func TestEngineWithIntegrations(t *testing.T) {
	if _, exists := os.LookupEnv("TESTS_ENABLE_INTEGRATIONS"); !exists {
		t.Skipf("TESTS_ENABLE_INTEGRATIONS disabled")
	}
	app := GetGinApp(ReadConfigFromEnvs())

	w := httptest.NewRecorder()
	inputComment := CommentModelInput{
		Author:       s("Test user Alex"),
		Email:        s("alex@example.com"),
		Website:      nil,
		Text:         "Hello, _world_",
		Parent:       nil,
		Title:        nil,
		Notification: 0,
	}
	inputCommentData, err := json.Marshal(inputComment)
	assert.Nil(t, err)
	req, _ := http.NewRequest(
		"POST",
		"/new?uri=example.com",
		strings.NewReader(string(inputCommentData)),
	)
	app.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	resultBody := strings.TrimSpace(w.Body.String())
	var resultModel CommentModelOutput

	json.Unmarshal([]byte(resultBody), &resultModel)

	assert.Equal(t, "Test user Alex", *resultModel.Author)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		"GET",
		"/?uri=example.com",
		nil,
	)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	allPageCommentsBody := strings.TrimSpace(w.Body.String())
	var pageCommentsData map[string]interface{}
	json.Unmarshal([]byte(allPageCommentsBody), &pageCommentsData)

	totalReplies, exists := pageCommentsData["total_replies"]
	assert.True(t, exists)
	totalRepliesTyped, err := getFloat(totalReplies)
	assert.Nil(t, err)
	assert.True(t, math.Abs(1-totalRepliesTyped) < 0.1)

	// allPageComments, exists := pageCommentsData["replies"]
	assert.True(t, exists)
}
