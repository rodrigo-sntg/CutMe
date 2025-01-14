package usecase_test

import (
	mocks2 "CutMe/internal/application/mocks"
	"CutMe/internal/application/usecase"
	"CutMe/internal/domain/entity"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessFileUseCase_Handle_DownloadError(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	s3Mock := new(mocks2.MockS3Client)
	dynamoMock := new(mocks2.MockDynamoClient)
	notifierMock := new(mocks2.MockNotifier)

	s3Mock.
		On("DownloadFile", "test-bucket", "sample.mp4").
		Return("", fmt.Errorf("arquivo não encontrado no S3"))

	dynamoMock.
		On("CreateOrUpdateUploadRecord", mock.AnythingOfType("domain.UploadRecord")).
		Return(nil).
		Once()

	notifierMock.On(
		"SendFailureEmail",
		"user123",
		"msgID123",
		mock.AnythingOfType("string"),
	).Return(nil)

	uc := usecase.NewProcessFileUseCase(
		s3Mock,
		"test-bucket",
		"cdn.test.com",
		dynamoMock,
		notifierMock,
	)

	ctx := context.Background()
	msg := entity.Message{
		ID:       "msgID123",
		FileName: "sample.mp4",
		UserID:   "user123",
	}

	err := uc.Handle(ctx, msg)
	assert.Error(t, err, "erro esperado ao falhar no download")

	s3Mock.AssertExpectations(t)
	dynamoMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}

func TestProcessFileUseCase_Handle_CreateRecordError(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	s3Mock := new(mocks2.MockS3Client)
	dynamoMock := new(mocks2.MockDynamoClient)
	notifierMock := new(mocks2.MockNotifier)

	dynamoMock.
		On("CreateOrUpdateUploadRecord", mock.AnythingOfType("domain.UploadRecord")).
		Return(fmt.Errorf("falha ao criar registro no DynamoDB"))

	uc := usecase.NewProcessFileUseCase(
		s3Mock,
		"test-bucket",
		"cdn.test.com",
		dynamoMock,
		notifierMock,
	)

	ctx := context.Background()
	msg := entity.Message{
		ID:       "msgID123",
		FileName: "sample.mp4",
		UserID:   "user123",
	}

	err := uc.Handle(ctx, msg)
	assert.Error(t, err, "erro esperado ao falhar na criação do registro no DynamoDB")

	dynamoMock.AssertExpectations(t)
	s3Mock.AssertNotCalled(t, "DownloadFile", mock.Anything, mock.Anything)
	notifierMock.AssertNotCalled(t, "SendFailureEmail", mock.Anything, mock.Anything, mock.Anything)
}

func TestProcessFileUseCase_Handle_UpdateRecordError(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	s3Mock := new(mocks2.MockS3Client)
	dynamoMock := new(mocks2.MockDynamoClient)
	notifierMock := new(mocks2.MockNotifier)

	localFile := "/tmp/test_video.mp4"
	s3Mock.On("DownloadFile", "test-bucket", "sample.mp4").Return(localFile, nil)

	notifierMock.On(
		"SendFailureEmail",
		"user123",
		"msgID123",
		mock.AnythingOfType("string"),
	).Return(nil)

	dynamoMock.
		On("CreateOrUpdateUploadRecord", mock.AnythingOfType("domain.UploadRecord")).
		Return(nil).
		Once()

	dynamoMock.
		On("CreateOrUpdateUploadRecord", mock.AnythingOfType("domain.UploadRecord")).
		Return(fmt.Errorf("falha ao atualizar registro no DynamoDB"))

	uc := usecase.NewProcessFileUseCase(
		s3Mock,
		"test-bucket",
		"cdn.test.com",
		dynamoMock,
		notifierMock,
	)

	ctx := context.Background()
	msg := entity.Message{
		ID:       "msgID123",
		FileName: "sample.mp4",
		UserID:   "user123",
	}

	err := uc.Handle(ctx, msg)

	assert.Error(t, err, "erro esperado ao extrair frames")
	assert.Contains(t, err.Error(), "ffmpeg-go", "erro esperado relacionado ao ffmpeg-go")

	s3Mock.AssertExpectations(t)
	dynamoMock.AssertExpectations(t)
	notifierMock.AssertExpectations(t)
}
