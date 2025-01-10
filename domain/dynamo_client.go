package domain

type DynamoClient interface {
	UpdateStatus(id, status string) error
	UpdateUploadRecord(record UploadRecord) error
	CreateUploadRecord(record UploadRecord) error
	CreateOrUpdateUploadRecord(record UploadRecord) error
	GetUploads(status string) ([]UploadRecord, error)
	GetUploadByID(id string) (*UploadRecord, error)
	GetUploadsByUserID(userID string, status string) ([]UploadRecord, error)
}
