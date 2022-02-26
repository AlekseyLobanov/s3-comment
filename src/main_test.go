package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
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

func getFakeInputComment() CommentModelInput {
	return CommentModelInput{
		Author:       s("Test user Alex"),
		Email:        s("alex@example.com"),
		Website:      nil,
		Text:         "Hello, _world_",
		Parent:       nil,
		Title:        nil,
		Notification: 0,
	}
}

func postDeleteS3Bucket(t *testing.T, config MinioConfig) {
	client := createMinioClient(&config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := client.ListObjects(ctx, config.Bucket, minio.ListObjectsOptions{
		Prefix:    "",
		Recursive: true,
	})
	for object := range objectCh {
		err := client.RemoveObject(ctx, config.Bucket, object.Key, minio.RemoveObjectOptions{})
		assert.Nil(t, err)
	}

	err := client.RemoveBucket(context.Background(), config.Bucket)

	assert.Nil(t, err)
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

func postComment(t *testing.T, app *gin.Engine, inputComment *CommentModelInput, uri string) CommentModelOutput {
	inputCommentData, err := json.Marshal(inputComment)
	assert.Nil(t, err)
	req, _ := http.NewRequest(
		"POST",
		"/new?uri="+uri,
		strings.NewReader(string(inputCommentData)),
	)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	resultBody := strings.TrimSpace(w.Body.String())
	var resultModel CommentModelOutput

	json.Unmarshal([]byte(resultBody), &resultModel)

	assert.Equal(t, *inputComment.Author, *resultModel.Author)
	return resultModel
}

func getCommentsForPage(t *testing.T,
	app *gin.Engine,
	uri string) int {

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"GET",
		"/?uri="+uri,
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
	return int(totalRepliesTyped)
}

func likeDislikeComment(t *testing.T, app *gin.Engine, commentId int64, action string) int {
	assert.True(t, action == "like" || action == "dislike")
	request_url := fmt.Sprintf("/id/%v/%v", commentId, action)
	req, _ := http.NewRequest(
		"POST",
		request_url,
		nil,
	)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w.Code
}

func TestEngineWithIntegrations(t *testing.T) {
	// really big single test for many things in one.
	// Not a great solution, but simple enough for good covearge/code ratio
	if _, exists := os.LookupEnv("TESTS_ENABLE_INTEGRATIONS"); !exists {
		t.Skipf("TESTS_ENABLE_INTEGRATIONS disabled")
	}
	testConfig := ReadConfigFromEnvs()
	testConfig.Minio.Bucket = "test"
	app := GetGinApp(testConfig)
	defer postDeleteS3Bucket(t, *testConfig.Minio)

	// naive negative check
	for _, action := range []string{"like", "dislike"} {
		assert.Equal(t, 422, likeDislikeComment(t, app, 41, action))
	}

	inputComment := getFakeInputComment()
	singleComment := postComment(t, app, &inputComment, "example.com/single")
	for _, action := range []string{"like", "dislike"} {
		assert.Equal(t, 200, likeDislikeComment(t, app, singleComment.Id, action))
	}
	postComment(t, app, &inputComment, "example.com/extra")
	postComment(t, app, &inputComment, "example.com/extra")

	for ind := 0; ind < 2; ind++ {
		singleCommentsCount := getCommentsForPage(t, app, "example.com/single")
		assert.Equal(t, 1, singleCommentsCount)
	}

	for ind := 0; ind < 2; ind++ {
		emptyCommentsCount := getCommentsForPage(t, app, "example.com/empty")
		assert.Equal(t, 0, emptyCommentsCount)
	}

	manyCommentsCount := getCommentsForPage(t, app, "example.com/extra")
	assert.Equal(t, 2, manyCommentsCount)
}
