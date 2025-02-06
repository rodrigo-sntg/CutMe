package sqs_test

import (
	"CutMe/internal/application/mocks"
	"CutMe/internal/domain/entity"
	sqs2 "CutMe/internal/infrastructure/aws/sqs"
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/mock/gomock"
)

func TestSQSConsumer_StartConsumption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueue := mocks.NewMockQueueClient(ctrl)
	mockStorage := mocks.NewMockStorageClient(ctrl)
	mockHandler := &mockMessageHandler{} // Simularemos manualmente

	queueURL := "https://sqs.amazonaws.com/123456789012/test-queue"
	consumer := sqs2.NewSQSConsumer(mockQueue, mockStorage, queueURL, mockHandler)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Run("fetch messages and process successfully", func(t *testing.T) {
		// Simula uma mensagem SQS válida
		messageBody := `{
			"Records": [
				{
					"s3": {
						"bucket": { "name": "test-bucket" },
						"object": { "key": "test-key" }
					}
				}
			]
		}`
		mockQueue.EXPECT().
			ReceiveMessage(gomock.Any()).
			Return(&sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{
						Body:          aws.String(messageBody),
						ReceiptHandle: aws.String("test-receipt-handle"),
					},
				},
			}, nil).AnyTimes()

		// Simula os metadados retornados pelo S3
		mockStorage.EXPECT().
			HeadObject("test-bucket", "test-key").
			Return(&s3.HeadObjectOutput{
				Metadata: map[string]*string{
					"Userid":    aws.String("123"),
					"Uniqueid":  aws.String("abc"),
					"Ignoresqs": aws.String("false"),
				},
			}, nil).AnyTimes()

		// Simula o processamento bem-sucedido do handler
		mockHandler.processed = false
		mockHandler.err = nil

		mockQueue.EXPECT().
			DeleteMessage(gomock.Any()).
			Return(&sqs.DeleteMessageOutput{}, nil).AnyTimes()

		consumer.StartConsumption(ctx, 1)
		if !mockHandler.processed {
			t.Fatalf("expected handler to process the message")
		}
	})

	t.Run("handle invalid JSON", func(t *testing.T) {
		invalidMessageBody := "invalid-json"
		mockQueue.EXPECT().
			ReceiveMessage(gomock.Any()).
			Return(&sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{
						Body:          aws.String(invalidMessageBody),
						ReceiptHandle: aws.String("test-receipt-handle"),
					},
				},
			}, nil).AnyTimes()

		mockQueue.EXPECT().
			DeleteMessage(gomock.Any()).
			Return(&sqs.DeleteMessageOutput{}, nil).AnyTimes()

		consumer.StartConsumption(ctx, 1)
		// Não espera erros neste caso
	})
}

// Mock do handler para testes
type mockMessageHandler struct {
	processed bool
	err       error
}

func (m *mockMessageHandler) Handle(ctx context.Context, msg entity.Message) error {
	m.processed = true
	return m.err
}
