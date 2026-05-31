package repository

import (
	"context"
	"evasbr/mclamg/entity"
)

type TransactionDetailRepository interface {
	FindById(ctx context.Context, id string) (entity.TransactionDetail, error)
}
