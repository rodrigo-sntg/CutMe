package service

import "context"

// QueueConsumer define os métodos necessários para consumir mensagens SQS.
type QueueConsumer interface {
	StartConsumption(ctx context.Context, workerCount int)
}
