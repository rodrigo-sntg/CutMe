package repository

import (
	"CutMe/internal/domain/entity"
)

type DBClient interface {
	UpdateStatus(id, status string) error
	UpdateUploadRecord(record entity.UploadRecord) error
	CreateUploadRecord(record entity.UploadRecord) error
	CreateOrUpdateUploadRecord(record entity.UploadRecord) error
	GetUploads(status string) ([]entity.UploadRecord, error)
	GetUploadByID(id string) (*entity.UploadRecord, error)
	GetUploadsByUserID(userID string, status string) ([]entity.UploadRecord, error)
}
