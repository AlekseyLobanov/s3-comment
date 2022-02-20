package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"

	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type S3CommentsBackend struct {
	minio            *minio.Client
	config           MinioConfig
	metricOperations *prometheus.CounterVec
}

func createMinioClient(config *MinioConfig) *minio.Client {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.Secure,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("New Minio: %#v\n", minioClient) // minioClient is now setup
	bucketExists, err := minioClient.BucketExists(context.Background(), config.Bucket)
	if err != nil {
		log.Fatalln(err)
	}
	if !bucketExists {
		log.Printf("Bucket %v not exists, creating", config.Bucket)
		err := minioClient.MakeBucket(
			context.Background(),
			config.Bucket,
			minio.MakeBucketOptions{
				Region:        "",
				ObjectLocking: false,
			})
		if err != nil {
			log.Fatalln(err)
		}
	}
	return minioClient
}

func NewS3CommentsStorage(config MinioConfig) (*S3CommentsBackend, error) {
	return &S3CommentsBackend{
		minio:  nil,
		config: config,
		metricOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "s3_requests",
			Help: "Number of S3 requests to comments storage",
		}, []string{"operation", "target"}),
	}, nil
	/*
		res := MemoryCommentsBackend{
			lastId:          2,
			commentsStorage: make(map[string][]int32),
			commentItems:    make(map[int32]*CommentModelOutput),
		}
		res.commentsStorage["/s3-comment.html"] = make([]int32, 0)
		res.commentsStorage["/s3-comment.html"] = append(
			res.commentsStorage["/s3-comment.html"],
			1,
		)
		res.commentsStorage["/s3-comment.html"] = append(
			res.commentsStorage["/s3-comment.html"],
			2,
		)
		res.commentItems[1] = &CommentModelOutput{
			Id: 1, Parent: nil, Created: 1642930664.2549465,
			Modified: nil, Mode: 1, Text: "<p>Hello, world (new)</p>", Author: s("Hippomoto"),
			Website: nil, Likes: 0, Dislikes: 0, Hash: "e4da2aacd5dc", TotalRelies: 0,
			HiddenReplies: 0, Replies: make([]CommentModelOutput, 0, 1),
		}
		res.commentItems[2] = &CommentModelOutput{
			Id: 2, Parent: nil, Created: 1642930664.2549465,
			Modified: nil, Mode: 1, Text: "<p>А этот <strong>посильнее</strong> </p>", Author: s("Blanket"),
			Website: nil, Likes: 0, Dislikes: 0, Hash: "e4da2aacd5dc", TotalRelies: 0,
			HiddenReplies: 0, Replies: make([]CommentModelOutput, 0, 1),
		}
		return &res, nil
	*/
}

func getCommetObjectName(commentId int64) string {
	return fmt.Sprintf("comments/%v.json", commentId)
}

func getUriObjectName(uri string) string {
	return fmt.Sprintf("pages/%v.json", CalculateUserHash(uri, "fakeTODO"))
}

func (backend *S3CommentsBackend) minioLazyInit() {
	if backend.minio == nil {
		backend.minio = createMinioClient(&backend.config)
	}
}

func (backend *S3CommentsBackend) saveCommentData(commentData *CommentModelOutput) error {
	backend.minioLazyInit()

	commentBytes, _ := json.Marshal(commentData)

	objectReader := bytes.NewReader(commentBytes)

	uploadInfo, err := backend.minio.PutObject(
		context.Background(),
		backend.config.Bucket,
		getCommetObjectName(commentData.Id),
		objectReader,
		int64(len(commentBytes)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	backend.metricOperations.WithLabelValues("PUT", "comment_data").Inc()
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)

	return nil
}

func (backend *S3CommentsBackend) GetPageComments(uri string) ([]int64, error) {
	backend.minioLazyInit()
	object, err := backend.minio.GetObject(
		context.Background(),
		backend.config.Bucket,
		getUriObjectName(uri),
		minio.GetObjectOptions{},
	)
	backend.metricOperations.WithLabelValues("GET", "page_comments").Inc()
	if err != nil {
		switch err.(type) {
		case minio.ErrorResponse:
			return make([]int64, 0), nil
		default:
			fmt.Println(err)
			return nil, err
		}
	}
	objectBytes, err := io.ReadAll(object)
	if err != nil {
		// срабатывает именно этот
		switch err.(type) {
		case minio.ErrorResponse:
			return make([]int64, 0), nil
		default:
			fmt.Println(err)
			return nil, err
		}
	}

	res := make([]int64, 0)
	err = json.Unmarshal(objectBytes, &res)
	if err != nil {
		log.Printf("Unable to load json with comments for page %v, error: %v\n", uri, err.Error())
		return nil, err
	}
	return res, nil
}

func (backend *S3CommentsBackend) AddCommentToPage(uri string, commentId int64) error {
	backend.minioLazyInit()
	currentComments, err := backend.GetPageComments(uri)
	if err != nil {
		log.Printf("Error %v with loading comments for page %v\n", err.Error(), uri)
		return errors.New("unable to load comments for page")
	}
	currentComments = append(currentComments, commentId)
	commentBytes, _ := json.Marshal(currentComments)

	objectReader := bytes.NewReader(commentBytes)

	uploadInfo, err := backend.minio.PutObject(
		context.Background(),
		backend.config.Bucket,
		getUriObjectName(uri),
		objectReader,
		int64(len(commentBytes)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	backend.metricOperations.WithLabelValues("PUT", "page_comments").Inc()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
	log.Printf("new comment_id: %v on page: %v\n", commentId, uri)

	return nil
}

func (backend *S3CommentsBackend) AddComment(commentData *CommentModelOutput) (int64, error) {
	error := backend.saveCommentData(commentData)
	return commentData.Id, error
}

func (backend *S3CommentsBackend) UpdateComment(commentData *CommentModelOutput) error {
	_, err := backend.AddComment(commentData)
	return err
}

func (backend *S3CommentsBackend) GetComment(commentId int64) (*CommentModelOutput, error) {
	backend.minioLazyInit()
	object, err := backend.minio.GetObject(
		context.Background(),
		backend.config.Bucket,
		getCommetObjectName(int64(commentId)),
		minio.GetObjectOptions{},
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	backend.metricOperations.WithLabelValues("GET", "comment_data").Inc()
	objectBytes, err := io.ReadAll(object)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res := CommentModelOutput{}
	err = json.Unmarshal(objectBytes, &res)
	if err != nil {
		log.Printf("Unable to load json with comment id %v, error: %v\n", commentId, err.Error())
		return nil, err
	}
	return &res, nil
}
