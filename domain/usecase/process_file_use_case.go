package usecase

import (
	"CutMe/domain/entity"
	"context"
)

type ProcessFileUseCase interface {
	Handle(ctx context.Context, msg entity.SQSMessage) error
}
