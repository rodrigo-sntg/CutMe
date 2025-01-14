package repository

import "github.com/aws/aws-sdk-go/service/sqs"

type QueueClient interface {
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
}
