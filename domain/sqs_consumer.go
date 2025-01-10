package domain

import "context"

// SQSConsumer define os métodos necessários para consumir mensagens SQS.
type SQSConsumer interface {
	StartConsumption(ctx context.Context, workerCount int)
}
