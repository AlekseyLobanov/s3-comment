package main

type CommentModelInput struct {
	Author       *string `json:"author"`
	Email        *string `json:"email"`
	Website      *string `json:"website"`
	Text         string  `json:"text"`
	Parent       *int64  `json:"parent"`
	Title        *string `json:"title"`
	Notification int     `json:"notification"`
}

type CommentModelOutput struct {
	Id            int64                `json:"id"`
	Parent        *int                 `json:"parent"`
	Created       float64              `json:"created"`
	Modified      *float64             `json:"modified"`
	Mode          int                  `json:"mode"`
	Text          string               `json:"text"`
	Author        *string              `json:"author"`
	Website       *string              `json:"website"`
	Likes         int                  `json:"likes"`
	Dislikes      int                  `json:"dislikes"`
	Notification  int                  `json:"notification"`
	Hash          string               `json:"hash"`
	TotalRelies   int                  `json:"total_replies"`
	HiddenReplies int                  `json:"hidden_replies"`
	Replies       []CommentModelOutput `json:"replies"`
}

type PreviewModel struct {
	// NB: Input model is equal to Output Model
	Text string `json:"text"`
}
