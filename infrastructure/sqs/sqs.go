package sqs

import (
	"CutMe/domain/entity"
	"CutMe/domain/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSConsumer struct {
	svc      repository.SQSAPI
	s3Client repository.S3Client
	queueURL string
	handler  entity.SQSMessageHandler
}

func NewSQSConsumer(svc repository.SQSAPI, s3Client repository.S3Client, queueURL string, handler entity.SQSMessageHandler) *SQSConsumer {
	return &SQSConsumer{
		svc:      svc,
		s3Client: s3Client,
		queueURL: queueURL,
		handler:  handler,
	}
}

func (c *SQSConsumer) StartConsumption(ctx context.Context, workerCount int) {
	log.Println("Iniciando consumo de mensagens SQS com workers...")
	var wg sync.WaitGroup
	messageChannel := make(chan *sqs.Message, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go c.worker(ctx, messageChannel, &wg)
	}

	go func() {
		defer close(messageChannel)
		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelado. Encerrando fetching.")
				return
			default:
				c.fetchMessages(ctx, messageChannel)
			}
		}
	}()

	wg.Wait()
	log.Println("Todos os workers finalizaram.")
}

func (c *SQSConsumer) fetchMessages(ctx context.Context, messageChannel chan<- *sqs.Message) {
	output, err := c.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: aws.Int64(5),
		WaitTimeSeconds:     aws.Int64(10),
	})
	if err != nil {
		log.Printf("Erro ao receber mensagem: %v\n", err)
		time.Sleep(5 * time.Second)
		return
	}

	for _, msg := range output.Messages {
		select {
		case <-ctx.Done():
			return
		case messageChannel <- msg:
		}
	}
}

func (c *SQSConsumer) worker(ctx context.Context, messageChannel <-chan *sqs.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range messageChannel {
		if msg.Body == nil {
			continue
		}

		var sqsEvent struct {
			Records []struct {
				S3 struct {
					Bucket struct {
						Name string `json:"name"`
					} `json:"bucket"`
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				} `json:"s3"`
			} `json:"Records"`
		}

		if err := json.Unmarshal([]byte(*msg.Body), &sqsEvent); err != nil {
			log.Printf("Erro ao decodificar evento S3: %v\n", err)
			c.deleteMessage(msg.ReceiptHandle)
			continue
		}

		for _, record := range sqsEvent.Records {
			bucketName := record.S3.Bucket.Name
			objectKey := record.S3.Object.Key

			metadata, err := c.fetchMetadataFromS3(bucketName, objectKey)
			if err != nil {
				log.Printf("Erro ao recuperar metadados: %v\n", err)
				continue
			}

			if metadata["Ignoresqs"] == "true" {
				log.Printf("Ignorando mensagem para arquivo: %s\n", objectKey)
				continue
			}

			userID := metadata["Userid"]
			uniqueID := metadata["Uniqueid"]

			sqsMsg := entity.SQSMessage{
				ID:              uniqueID,
				FileName:        objectKey,
				Bucket:          bucketName,
				UserID:          userID,
				OriginalFileURL: fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, objectKey),
				UploadedAt:      time.Now().Unix(),
			}

			if err := c.handler.Handle(ctx, sqsMsg); err != nil {
				log.Printf("Erro ao processar mensagem: %v\n", err)
			}

			c.deleteMessage(msg.ReceiptHandle)
		}
	}
}

func (c *SQSConsumer) fetchMetadataFromS3(bucketName, objectKey string) (map[string]string, error) {
	headOutput, err := c.s3Client.HeadObject(bucketName, objectKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter metadados do S3: %w", err)
	}

	metadata := make(map[string]string)
	for key, value := range headOutput.Metadata {
		metadata[key] = aws.StringValue(value)
	}
	return metadata, nil
}

func (c *SQSConsumer) deleteMessage(receiptHandle *string) {
	_, err := c.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("Falha ao deletar mensagem da fila: %v\n", err)
	}
}
