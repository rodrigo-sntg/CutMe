package domain

import (
	"context"
)

type ProcessFileUseCase interface {
	Handle(ctx context.Context, msg SQSMessage) error
}
