package repository

import "time"

// PresignedURLGenerator define a interface para geração de URLs assinadas
type PresignedURLGenerator interface {
	GeneratePreSignedURL(fileName, fileType, userID string) (signedURL string, uniqueID string, err error)
	GetBucketName() string
	GetURLValidity() time.Duration
}
