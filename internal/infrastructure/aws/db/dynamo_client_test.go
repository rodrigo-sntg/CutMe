package db

import (
	"CutMe/internal/application/mocks"
	"CutMe/internal/domain/entity"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
)

func TestUpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(&dynamodb.UpdateItemOutput{}, nil)

		err := client.UpdateStatus("123", "PROCESSING")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error updating item", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(nil, errors.New("update error"))

		err := client.UpdateStatus("123", "FAILED")
		if err == nil || err.Error() != "erro ao atualizar status no Dynamo: update error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateUploadRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	record := entity.UploadRecord{
		ID:               "123",
		UserID:           "user-1",
		OriginalFileName: "file.txt",
		UniqueFileName:   "unique-file.txt",
		Status:           "CREATED",
		CreatedAt:        1620000000,
		OriginalFileURL:  "http://example.com/file.txt",
		ProcessedFileURL: "http://example.com/processed-file.txt",
	}

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().PutItem(gomock.Any()).Return(&dynamodb.PutItemOutput{}, nil)

		err := client.CreateUploadRecord(record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("conditional check failed", func(t *testing.T) {
		mockDynamo.EXPECT().
			PutItem(gomock.Any()).
			Return(nil, fmt.Errorf(dynamodb.ErrCodeConditionalCheckFailedException))

		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(&dynamodb.UpdateItemOutput{}, nil)

		err := client.CreateUploadRecord(record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error putting item", func(t *testing.T) {
		mockDynamo.EXPECT().PutItem(gomock.Any()).Return(nil, errors.New("put error"))

		err := client.CreateUploadRecord(record)
		if err == nil || err.Error() != "erro ao criar registro no DynamoDB: put error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetUploadByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().
			GetItem(gomock.Any()).
			Return(&dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"id": {
						S: aws.String("123"),
					},
					"status": {
						S: aws.String("PROCESSING"),
					},
				},
			}, nil)

		result, err := client.GetUploadByID("123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != "123" || result.Status != "PROCESSING" {
			t.Fatalf("unexpected result: %v", result)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		mockDynamo.EXPECT().
			GetItem(gomock.Any()).
			Return(&dynamodb.GetItemOutput{Item: nil}, nil)

		_, err := client.GetUploadByID("123")
		if err == nil || err.Error() != "registro com ID 123 não encontrado" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error getting item", func(t *testing.T) {
		mockDynamo.EXPECT().
			GetItem(gomock.Any()).
			Return(nil, errors.New("get error"))

		_, err := client.GetUploadByID("123")
		if err == nil || err.Error() != "erro ao buscar registro no DynamoDB: get error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetUploads(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	t.Run("success with filter", func(t *testing.T) {
		mockDynamo.EXPECT().
			Scan(gomock.Any()).
			Return(&dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{
						"id": {
							S: aws.String("123"),
						},
						"status": {
							S: aws.String("PROCESSING"),
						},
					},
				},
			}, nil)

		results, err := client.GetUploads("PROCESSING")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 || results[0].ID != "123" {
			t.Fatalf("unexpected results: %v", results)
		}
	})

	t.Run("error scanning items", func(t *testing.T) {
		mockDynamo.EXPECT().
			Scan(gomock.Any()).
			Return(nil, errors.New("scan error"))

		_, err := client.GetUploads("PROCESSING")
		if err == nil || err.Error() != "erro ao escanear registros no DynamoDB: scan error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateUploadRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	record := entity.UploadRecord{
		ID:             "123",
		UniqueFileName: "updated-file.txt",
		Status:         "PROCESSED",
		ProcessedAt:    1620000000,
	}

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(&dynamodb.UpdateItemOutput{}, nil)

		err := client.UpdateUploadRecord(record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(nil, errors.New("update error"))

		err := client.UpdateUploadRecord(record)
		if err == nil || err.Error() != "erro ao atualizar registro no DynamoDB: update error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateOrUpdateUploadRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	record := entity.UploadRecord{
		ID:               "123",
		OriginalFileName: "file.txt",
		UniqueFileName:   "unique-file.txt",
		Status:           "CREATED",
		CreatedAt:        1620000000,
		ProcessedAt:      1620000001,
	}

	t.Run("success", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(&dynamodb.UpdateItemOutput{}, nil)

		err := client.CreateOrUpdateUploadRecord(record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update error", func(t *testing.T) {
		mockDynamo.EXPECT().UpdateItem(gomock.Any()).Return(nil, errors.New("update error"))

		err := client.CreateOrUpdateUploadRecord(record)
		if err == nil || err.Error() != "erro ao criar ou atualizar registro no DynamoDB: update error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetUploadsByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamo := mocks.NewMockDynamoDBAPI(ctrl)
	client := &dynamoClient{Svc: mockDynamo, TableName: "test-table"}

	t.Run("success without status filter", func(t *testing.T) {
		mockDynamo.EXPECT().Query(gomock.Any()).Return(&dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{
					"id": {
						S: aws.String("123"),
					},
					"userId": {
						S: aws.String("user-1"),
					},
				},
			},
		}, nil)

		results, err := client.GetUploadsByUserID("user-1", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 || results[0].ID != "123" {
			t.Fatalf("unexpected results: %v", results)
		}
	})

	t.Run("error querying items", func(t *testing.T) {
		mockDynamo.EXPECT().Query(gomock.Any()).Return(nil, errors.New("query error"))

		_, err := client.GetUploadsByUserID("user-1", "")
		if err == nil || err.Error() != "erro ao consultar registros no DynamoDB: query error" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewDynamoClientWithRealSession(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		t.Fatalf("failed to create AWS session: %v", err)
	}

	client := NewDynamoClient(sess, "test-table")

	if client == nil {
		t.Fatal("DynamoDB client should not be nil")
	}

	dynamoClient, ok := client.(*dynamoClient)
	if !ok || dynamoClient.TableName != "test-table" || dynamoClient.Svc == nil {
		t.Fatal("unexpected client initialization")
	}
}
