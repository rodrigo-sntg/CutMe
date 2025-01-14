package repository

import "time"

// SignedURLGenerator define a interface para geração de URLs assinadas
type SignedURLGenerator interface {
	GeneratePreSignedURL(fileName, fileType, userID string) (signedURL string, uniqueID string, err error)
	GetBucketName() string
	GetURLValidity() time.Duration
}
