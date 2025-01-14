package usecase

import (
	"CutMe/internal/domain/entity"
	"context"
)

type ProcessFileUseCase interface {
	Handle(ctx context.Context, msg entity.Message) error
}
