package impl

import (
	"context"
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func NewProductServiceImpl(productRepository *repository.ProductRepository, cache *redis.Client) service.ProductService {
	return &productServiceImpl{
		ProductRepository: *productRepository,
		Cache:             cache,
		log:               common.Log.WithField("scope", "ProductService"),
	}
}

type productServiceImpl struct {
	repository.ProductRepository
	Cache *redis.Client
	log   *logrus.Entry
}

func (service *productServiceImpl) Create(ctx context.Context, productModel model.ProductCreateOrUpdateModel) model.ProductCreateOrUpdateModel {
	service.log.WithContext(ctx).Info("Creating product: ", productModel.Name)
	common.Validate(productModel)
	product := entity.Product{
		Name:     productModel.Name,
		Price:    productModel.Price,
		Quantity: productModel.Quantity,
	}
	service.ProductRepository.Insert(ctx, product)
	return productModel
}

func (service *productServiceImpl) Update(ctx context.Context, productModel model.ProductCreateOrUpdateModel, id string) model.ProductCreateOrUpdateModel {
	common.Validate(productModel)
	product := entity.Product{
		Id:       uuid.MustParse(id),
		Name:     productModel.Name,
		Price:    productModel.Price,
		Quantity: productModel.Quantity,
	}
	service.ProductRepository.Update(ctx, product)
	return productModel
}

func (service *productServiceImpl) Delete(ctx context.Context, id string) {
	product, err := service.ProductRepository.FindById(ctx, id)
	if err != nil {
		panic(exception.NotFoundError{
			Message: err.Error(),
		})
	}
	service.ProductRepository.Delete(ctx, product)
}

func (service *productServiceImpl) FindById(ctx context.Context, id string) model.ProductModel {
	productCache := configuration.SetCache[entity.Product](service.Cache, ctx, "product", id, service.ProductRepository.FindById)
	return model.ProductModel{
		Id:       productCache.Id.String(),
		Name:     productCache.Name,
		Price:    productCache.Price,
		Quantity: productCache.Quantity,
	}
}

func (service *productServiceImpl) FindAll(ctx context.Context) (responses []model.ProductModel) {
	products := service.ProductRepository.FindAl(ctx)
	for _, product := range products {
		responses = append(responses, model.ProductModel{
			Id:       product.Id.String(),
			Name:     product.Name,
			Price:    product.Price,
			Quantity: product.Quantity,
		})
	}
	if len(products) == 0 {
		return []model.ProductModel{}
	}
	return responses
}
