package service

import (
	"context"
	"evasbr/mclamg/model"
)

type TransactionDetailService interface {
	FindById(ctx context.Context, id string) model.TransactionDetailModel
}
