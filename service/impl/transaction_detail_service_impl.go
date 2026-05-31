package impl

import (
	"context"
	"evasbr/mclamg/common"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"
	"github.com/sirupsen/logrus"
)

func NewTransactionDetailServiceImpl(transactionDetailRepository *repository.TransactionDetailRepository) service.TransactionDetailService {
	return &transactionDetailServiceImpl{
		TransactionDetailRepository: *transactionDetailRepository,
		log:                         common.Log.WithField("scope", "TransactionDetailService"),
	}
}

type transactionDetailServiceImpl struct {
	repository.TransactionDetailRepository
	log *logrus.Entry
}

func (transactionDetailService *transactionDetailServiceImpl) FindById(ctx context.Context, id string) model.TransactionDetailModel {
	transactionDetail, err := transactionDetailService.TransactionDetailRepository.FindById(ctx, id)
	if err != nil {
		panic(exception.NotFoundError{
			Message: err.Error(),
		})
	}
	return model.TransactionDetailModel{
		Id:            transactionDetail.Id.String(),
		SubTotalPrice: transactionDetail.SubTotalPrice,
		Price:         transactionDetail.Price,
		Quantity:      transactionDetail.Quantity,
		Product: model.ProductModel{
			Id:       transactionDetail.Product.Id.String(),
			Name:     transactionDetail.Product.Name,
			Price:    transactionDetail.Product.Price,
			Quantity: transactionDetail.Product.Quantity,
		},
	}
}
