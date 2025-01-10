package domain

import "github.com/aws/aws-sdk-go/service/s3"

type S3Client interface {
	DownloadFile(bucket, key string) (string, error)
	UploadFile(bucket, key, localPath string) error
	HeadObject(bucket, key string) (*s3.HeadObjectOutput, error) // Alterado para aceitar strings diretamente

}
