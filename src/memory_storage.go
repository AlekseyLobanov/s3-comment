package main

import "fmt"

type MemoryCommentsStorageLinked struct {
	commentItems    map[int64]*CommentModelOutput
	commentsStorage map[string][]int64
	slowBackend     CommentsStorageInterface
}

func NewMemoryStorageLinked(slowBackend CommentsStorageInterface) (*MemoryCommentsStorageLinked, error) {
	return &MemoryCommentsStorageLinked{
		commentItems:    make(map[int64]*CommentModelOutput),
		commentsStorage: make(map[string][]int64),
		slowBackend:     slowBackend,
	}, nil
}

func (storage *MemoryCommentsStorageLinked) GetPageComments(uri string) ([]int64, error) {
	value, exists := storage.commentsStorage[uri]
	if exists {
		return value, nil
	}
	if storage.slowBackend == nil {
		return nil, fmt.Errorf("uri %v not found, but slowBackend is not available", uri)
	}
	value, error := storage.slowBackend.GetPageComments(uri)
	if error != nil {
		return value, error
	}
	storage.commentsStorage[uri] = value
	return value, nil
}

func (storage *MemoryCommentsStorageLinked) AddCommentToPage(uri string, commentId int64) error {
	if storage.slowBackend != nil {
		err := storage.slowBackend.AddCommentToPage(uri, commentId)
		if err != nil {
			return err
		}
	}

	_, exists := storage.commentsStorage[uri]
	if !exists {
		storage.commentsStorage[uri] = make([]int64, 0)
	}
	storage.commentsStorage[uri] = append(storage.commentsStorage[uri], commentId)
	return nil
}

func (storage *MemoryCommentsStorageLinked) putComment(commentData *CommentModelOutput) error {
	storage.commentItems[commentData.Id] = commentData
	return nil
}

func (storage *MemoryCommentsStorageLinked) AddComment(commentData *CommentModelOutput) (int64, error) {
	commentId := commentData.Id
	if storage.slowBackend != nil {
		commentId, err := storage.slowBackend.AddComment(commentData)
		if err != nil {
			return 0, err
		}
		storage.putComment(commentData)
		return commentId, nil
	}
	storage.putComment(commentData)
	return commentId, nil
}

func (storage *MemoryCommentsStorageLinked) UpdateComment(commentData *CommentModelOutput) error {
	if storage.slowBackend != nil {
		err := storage.slowBackend.UpdateComment(commentData)
		if err != nil {
			return err
		}
	}
	storage.putComment(commentData)
	return nil
}

func (storage *MemoryCommentsStorageLinked) GetComment(commentId int64) (*CommentModelOutput, error) {
	value, exists := storage.commentItems[commentId]
	if exists {
		return value, nil
	}
	if storage.slowBackend == nil {
		return nil, fmt.Errorf("comment %v not found, but slowBackend is not available", commentId)
	}
	value, error := storage.slowBackend.GetComment(commentId)
	if error != nil {
		return value, error
	}
	storage.commentItems[commentId] = value
	return value, nil
}
