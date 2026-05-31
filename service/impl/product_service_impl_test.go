package impl

import (
	"context"
	"errors"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// MockProductRepository is a mock implementation of repository.ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Insert(ctx context.Context, product entity.Product) entity.Product {
	args := m.Called(ctx, product)
	return args.Get(0).(entity.Product)
}

func (m *MockProductRepository) Update(ctx context.Context, product entity.Product) entity.Product {
	args := m.Called(ctx, product)
	return args.Get(0).(entity.Product)
}

func (m *MockProductRepository) Delete(ctx context.Context, product entity.Product) {
	m.Called(ctx, product)
}

func (m *MockProductRepository) FindById(ctx context.Context, id string) (entity.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindAl(ctx context.Context) []entity.Product {
	args := m.Called(ctx)
	return args.Get(0).([]entity.Product)
}

func TestProductService_Create(t *testing.T) {
	mockRepo := new(MockProductRepository)
	var repo repository.ProductRepository = mockRepo
	service := NewProductServiceImpl(&repo, nil)

	ctx := context.Background()
	req := model.ProductCreateOrUpdateModel{
		Name:     "Milo Premium",
		Price:    7500,
		Quantity: 50,
	}

	expectedProduct := entity.Product{
		Name:     "Milo Premium",
		Price:    7500,
		Quantity: 50,
	}

	mockRepo.On("Insert", ctx, mock.MatchedBy(func(p entity.Product) bool {
		return p.Name == expectedProduct.Name && p.Price == expectedProduct.Price && p.Quantity == expectedProduct.Quantity
	})).Return(expectedProduct)

	res := service.Create(ctx, req)

	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Price, res.Price)
	assert.Equal(t, req.Quantity, res.Quantity)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_Success(t *testing.T) {
	mockRepo := new(MockProductRepository)
	var repo repository.ProductRepository = mockRepo
	service := NewProductServiceImpl(&repo, nil)

	ctx := context.Background()
	id := uuid.New().String()
	foundProduct := entity.Product{
		Id:       uuid.MustParse(id),
		Name:     "Milo Premium",
		Price:    7500,
		Quantity: 50,
	}

	mockRepo.On("FindById", ctx, id).Return(foundProduct, nil)
	mockRepo.On("Delete", ctx, foundProduct).Return()

	assert.NotPanics(t, func() {
		service.Delete(ctx, id)
	})

	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockProductRepository)
	var repo repository.ProductRepository = mockRepo
	service := NewProductServiceImpl(&repo, nil)

	ctx := context.Background()
	id := uuid.New().String()

	mockRepo.On("FindById", ctx, id).Return(entity.Product{}, errors.New("product Not Found"))

	assert.Panics(t, func() {
		service.Delete(ctx, id)
	})

	mockRepo.AssertExpectations(t)
}
