package main

type CommentsStorageInterface interface {
	GetPageComments(uri string) ([]int64, error)
	AddCommentToPage(uri string, commentId int64) error
	AddComment(commentData *CommentModelOutput) (int64, error) // comment id
	UpdateComment(commentData *CommentModelOutput) error
	GetComment(commentId int64) (*CommentModelOutput, error)
}

func likeModifier(comment *CommentModelOutput) {
	comment.Likes += 1
}

func dislikeModifier(comment *CommentModelOutput) {
	comment.Dislikes += 1
}
