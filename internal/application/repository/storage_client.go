package repository

import "github.com/aws/aws-sdk-go/service/s3"

type StorageClient interface {
	DownloadFile(bucket, key string) (string, error)
	UploadFile(bucket, key, localPath string) error
	HeadObject(bucket, key string) (*s3.HeadObjectOutput, error)
}
