package main

import (
	"fmt"
	"log"
	"time"
)

type CommentsLogicInterface interface {
	AddComment(uri string, inputComment *CommentModelInput) (*CommentModelOutput, error)
	GetComments(uri string, nestedLimit int) []*CommentModelOutput
	Like(commentId int64) (int64, int64, error)
	Dislike(commentId int64) (int64, int64, error)
}

type SimpleCommentsLogic struct {
	storageS3     CommentsStorageInterface
	storageMemory CommentsStorageInterface
	storage       CommentsStorageInterface
}

func GetCommentsLogic() *SimpleCommentsLogic {
	storageS3, err := NewS3CommentsStorage()
	storageMemory, _ := NewMemoryStorageLinked(storageS3)
	if err != nil {
		log.Fatalf("Unable to init comments storage, error: %v", err.Error())
	}
	return &SimpleCommentsLogic{
		storageS3:     storageS3,
		storageMemory: storageMemory,
		storage:       storageMemory,
	}
}

func (logic *SimpleCommentsLogic) AddComment(uri string, inputComment *CommentModelInput) (*CommentModelOutput, error) {
	if inputComment.Parent != nil {
		parentComment, _ := logic.storage.GetComment(*inputComment.Parent)
		if parentComment == nil {
			return nil, fmt.Errorf("parent comment id: %v is unknown", *inputComment.Parent)
		}
	}
	newId := time.Now().UnixMilli()

	res := CommentModelOutput{
		Id:            newId,
		Parent:        nil,
		Created:       float64(time.Now().UnixMilli()) / 1000,
		Modified:      nil,
		Mode:          1,
		Text:          RenderMarkdown(inputComment.Text),
		Author:        inputComment.Author,
		Website:       inputComment.Website,
		Likes:         0,
		Dislikes:      0,
		Notification:  1,
		Hash:          CalculateUserHash(*inputComment.Email, "SECRET_KEY"),
		TotalRelies:   0,
		HiddenReplies: 0,
		Replies:       []CommentModelOutput{},
	}
	_, err := logic.storage.AddComment(&res)
	if err != nil {
		log.Printf("Unable to add comment to storage: %v\n", err.Error())
		return nil, err
	}
	err = logic.storage.AddCommentToPage(uri, res.Id)
	if err != nil {
		log.Printf("Unable to add comment %v to page %v in storage, eror: %v\n", res.Id, uri, err.Error())
		return nil, err
	}
	log.Printf("new comment: %v\n", res)
	return &res, nil
}

func (logic *SimpleCommentsLogic) GetComments(uri string, nestedLimit int) []*CommentModelOutput {
	commentIds, error := logic.storage.GetPageComments(uri)
	if error != nil {
		log.Printf("Unable to load comments for %v\n", uri)
		return make([]*CommentModelOutput, 0)
	}

	res := make([]*CommentModelOutput, 0, len(commentIds))
	for _, comment := range commentIds {
		commentData, err := logic.storage.GetComment(int64(comment))
		if commentData == nil || err != nil {
			log.Printf("Unable to load comment %v from page %v\n", comment, uri)
			continue
		}
		res = append(res, commentData)
	}
	return res
}

func likeDislikeProcessorLogic(
	logic *SimpleCommentsLogic,
	commentId int64,
	modifier func(*CommentModelOutput),
) (int64, int64, error) {
	comment, error := logic.storage.GetComment(commentId)
	if error != nil {
		return 0, 0, error
	}
	if comment == nil {
		return 0, 0, fmt.Errorf("comment with id: %v not found", commentId)
	}
	modifier(comment)
	error = logic.storage.UpdateComment(comment)
	if error != nil {
		return 0, 0, error
	}
	return int64(comment.Likes), int64(comment.Dislikes), nil
}

func (logic *SimpleCommentsLogic) Like(commentId int64) (int64, int64, error) {
	return likeDislikeProcessorLogic(logic, commentId, likeModifier)
}

func (logic *SimpleCommentsLogic) Dislike(commentId int64) (int64, int64, error) {
	return likeDislikeProcessorLogic(logic, commentId, dislikeModifier)
}
