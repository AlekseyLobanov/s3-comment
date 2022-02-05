package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const APPLICATION_PORT = 8123

func likeDislikeHandler(c *gin.Context, backendHandler func(int64) (int64, int64, error)) {
	commentId, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		c.PureJSON(http.StatusUnprocessableEntity, gin.H{
			"error": fmt.Sprintf("Invalid commentId: %v", c.Param("commentId")),
		})
	}
	likes, dislikes, isOk := backendHandler(
		int64(commentId))
	if isOk != nil {
		c.PureJSON(http.StatusUnprocessableEntity, gin.H{
			"likes":    likes,
			"dislikes": dislikes,
			"error":    isOk.Error(),
		})
		return
	}
	c.PureJSON(200, gin.H{
		"likes":    likes,
		"dislikes": dislikes,
	})
}

func GetGinApp() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://127.0.0.1:8800"},
		AllowMethods: []string{http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "X-XSRF-TOKEN", "Accept", "Origin", "X-Requested-With", "Authorization"},
		ExposeHeaders: []string{
			"Content-Length",
			"Date", // client error without this line, in timezone calculation
		},
		AllowCredentials: true,
	}))

	metricsMonitor := GetPrometheusHandler()
	metricsMonitor.Use(r)

	// commentsBackend, _ := NewMemoryCommentsStorage()
	commentsBackend := GetCommentsLogic()

	r.Static("/js", "../static/js")
	r.Static("/css", "../static/css")

	r.OPTIONS("/count", func(c *gin.Context) {
		c.String(200, "")
	})
	r.POST("/count", func(c *gin.Context) {
		c.PureJSON(200, make([]int, 0))
	})

	r.OPTIONS("/", func(c *gin.Context) {
		c.String(200, "")
	})
	r.GET("/", func(c *gin.Context) {
		uri := c.Query("uri")
		if len(uri) == 0 {
			c.PureJSON(http.StatusUnprocessableEntity, gin.H{
				"error": "No uri in query",
			})
		}
		log.Printf("Comments request for %v\n", uri)
		comments := commentsBackend.GetComments(uri, 10)
		c.PureJSON(200, gin.H{
			"id":             nil,
			"total_replies":  len(comments),
			"hidden_replies": 0,
			"replies":        comments,
		})
	})

	r.POST("/preview", func(c *gin.Context) {
		inputComment := PreviewModel{}
		if err := c.ShouldBindJSON(&inputComment); err != nil {
			c.PureJSON(http.StatusUnprocessableEntity, "Invalid input model")
			log.Printf("Preview error: %v\n", err.Error())
			return
		}
		outputComment := PreviewModel{Text: RenderMarkdown(inputComment.Text)}
		c.PureJSON(200, outputComment)
	})
	r.POST("/new", func(c *gin.Context) {
		uri := c.Query("uri")
		if len(uri) == 0 {
			c.PureJSON(http.StatusUnprocessableEntity, gin.H{
				"error": "No uri in query",
			})
			return
		}
		inputComment := CommentModelInput{}
		if err := c.ShouldBindJSON(&inputComment); err != nil {
			c.PureJSON(http.StatusUnprocessableEntity, gin.H{
				"error": "Invalid input model",
			})
			return
		}
		newComment, err := commentsBackend.AddComment(uri, &inputComment)
		if err != nil {
			c.PureJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.PureJSON(201, newComment)
	})
	r.POST("/id/:commentId/like", func(c *gin.Context) {
		likeDislikeHandler(c, func(commentId int64) (int64, int64, error) {
			return commentsBackend.Like(commentId)
		})
	})
	r.POST("/id/:commentId/dislike", func(c *gin.Context) {
		likeDislikeHandler(c, func(commentId int64) (int64, int64, error) {
			return commentsBackend.Dislike(commentId)
		})
	})
	return r
}

func main() {
	fmt.Printf("s3-comment, builded with Go %s\n", runtime.Version())

	app := GetGinApp()
	app.Run("0.0.0.0:" + strconv.Itoa(APPLICATION_PORT))
}
