package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/asssoygo/pharmacy-product-service/internal/domain"
	"github.com/asssoygo/pharmacy-product-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mocks ---

type mockProductRepo struct{ mock.Mock }

func (m *mockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}
func (m *mockProductRepo) List(ctx context.Context, page, limit int) ([]*domain.Product, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*domain.Product), args.Int(1), args.Error(2)
}
func (m *mockProductRepo) Update(ctx context.Context, p *domain.Product) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProductRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockProductRepo) SearchByNameOrCategory(ctx context.Context, q string, page, limit int) ([]*domain.Product, int, error) {
	args := m.Called(ctx, q, page, limit)
	return args.Get(0).([]*domain.Product), args.Int(1), args.Error(2)
}
func (m *mockProductRepo) GetLowStock(ctx context.Context) ([]*domain.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Product), args.Error(1)
}
func (m *mockProductRepo) GetExpired(ctx context.Context) ([]*domain.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Product), args.Error(1)
}
func (m *mockProductRepo) GetByCategory(ctx context.Context, catID string, page, limit int) ([]*domain.Product, int, error) {
	args := m.Called(ctx, catID, page, limit)
	return args.Get(0).([]*domain.Product), args.Int(1), args.Error(2)
}
func (m *mockProductRepo) UpdateStock(ctx context.Context, id string, qty int32) (*domain.Product, error) {
	args := m.Called(ctx, id, qty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

type mockCategoryRepo struct{ mock.Mock }

func (m *mockCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCategoryRepo) GetAll(ctx context.Context) ([]*domain.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Category), args.Error(1)
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

// --- tests ---

func TestCreateProduct_Success(t *testing.T) {
	pr := new(mockProductRepo)
	cr := new(mockCategoryRepo)
	uc := usecase.NewProductUseCase(pr, cr)

	input := &domain.Product{
		Name:        "Aspirin",
		Description: "Pain reliever",
		Price:       5.99,
		Quantity:    100,
		CategoryID:  "cat-1",
	}
	pr.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == "Aspirin" && p.ID != ""
	})).Return(nil)

	result, err := uc.CreateProduct(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Aspirin", result.Name)
	assert.False(t, result.CreatedAt.IsZero())
	pr.AssertExpectations(t)
}

func TestGetProduct_NotFound(t *testing.T) {
	pr := new(mockProductRepo)
	cr := new(mockCategoryRepo)
	uc := usecase.NewProductUseCase(pr, cr)

	pr.On("GetByID", mock.Anything, "missing-id").Return(nil, domain.ErrProductNotFound)

	result, err := uc.GetProduct(context.Background(), "missing-id")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrProductNotFound))
	pr.AssertExpectations(t)
}

func TestUpdateStock_InsufficientStock(t *testing.T) {
	pr := new(mockProductRepo)
	cr := new(mockCategoryRepo)
	uc := usecase.NewProductUseCase(pr, cr)

	result, err := uc.UpdateStock(context.Background(), "prod-1", -5)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrInsufficientStock))
	pr.AssertNotCalled(t, "UpdateStock")
}

func TestUpdateStock_Success(t *testing.T) {
	pr := new(mockProductRepo)
	cr := new(mockCategoryRepo)
	uc := usecase.NewProductUseCase(pr, cr)

	expected := &domain.Product{ID: "prod-1", Name: "Aspirin", Quantity: 50, UpdatedAt: time.Now()}
	pr.On("UpdateStock", mock.Anything, "prod-1", int32(50)).Return(expected, nil)

	result, err := uc.UpdateStock(context.Background(), "prod-1", 50)

	assert.NoError(t, err)
	assert.Equal(t, int32(50), result.Quantity)
	pr.AssertExpectations(t)
}

func TestDeleteProduct_NotFound(t *testing.T) {
	pr := new(mockProductRepo)
	cr := new(mockCategoryRepo)
	uc := usecase.NewProductUseCase(pr, cr)

	pr.On("GetByID", mock.Anything, "ghost").Return(nil, domain.ErrProductNotFound)

	err := uc.DeleteProduct(context.Background(), "ghost")

	assert.True(t, errors.Is(err, domain.ErrProductNotFound))
	pr.AssertNotCalled(t, "Delete")
}
