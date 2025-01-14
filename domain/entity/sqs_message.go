package entity

import "context"

type SQSMessageHandler interface {
	Handle(ctx context.Context, msg SQSMessage) error
}

type SQSMessage struct {
	ID              string `json:"id"`
	FileName        string `json:"fileName"`
	Bucket          string `json:"bucket"`
	UniqueID        string `json:"UniqueID"`
	UserID          string `json:"userID"`
	OriginalFileURL string `json:"originalFileUrl"`
	UploadedAt      int64  `json:"uploadedAt"`
}
