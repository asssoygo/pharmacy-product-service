package usecase

import (
	"context"
	"time"

	"github.com/asssoygo/pharmacy-product-service/internal/domain"
	"github.com/google/uuid"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	GetProducts(ctx context.Context, page, limit int) ([]*domain.Product, int, error)
	UpdateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	SearchProducts(ctx context.Context, query string, page, limit int) ([]*domain.Product, int, error)
	GetLowStockProducts(ctx context.Context) ([]*domain.Product, error)
	UpdateStock(ctx context.Context, id string, quantity int32) (*domain.Product, error)
	GetExpiredProducts(ctx context.Context) ([]*domain.Product, error)
	GetProductsByCategory(ctx context.Context, categoryID string, page, limit int) ([]*domain.Product, int, error)
	CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error)
	GetCategories(ctx context.Context) ([]*domain.Category, error)
}

type productUseCase struct {
	productRepo  domain.ProductRepository
	categoryRepo domain.CategoryRepository
}

func NewProductUseCase(pr domain.ProductRepository, cr domain.CategoryRepository) ProductUseCase {
	return &productUseCase{
		productRepo:  pr,
		categoryRepo: cr,
	}
}

func (u *productUseCase) CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	p.ID = uuid.NewString()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	if err := u.productRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (u *productUseCase) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return u.productRepo.GetByID(ctx, id)
}

func (u *productUseCase) GetProducts(ctx context.Context, page, limit int) ([]*domain.Product, int, error) {
	return u.productRepo.List(ctx, page, limit)
}

func (u *productUseCase) UpdateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	existing, err := u.productRepo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if p.Name != "" {
		existing.Name = p.Name
	}
	if p.Description != "" {
		existing.Description = p.Description
	}
	if p.Price > 0 {
		existing.Price = p.Price
	}
	if p.CategoryID != "" {
		existing.CategoryID = p.CategoryID
	}
	if p.MinStockLevel >= 0 {
		existing.MinStockLevel = p.MinStockLevel
	}
	if p.ExpiryDate != nil {
		existing.ExpiryDate = p.ExpiryDate
	}
	existing.UpdatedAt = time.Now().UTC()
	if err := u.productRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *productUseCase) DeleteProduct(ctx context.Context, id string) error {
	if _, err := u.productRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return u.productRepo.Delete(ctx, id)
}

func (u *productUseCase) SearchProducts(ctx context.Context, query string, page, limit int) ([]*domain.Product, int, error) {
	return u.productRepo.SearchByNameOrCategory(ctx, query, page, limit)
}

func (u *productUseCase) GetLowStockProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.productRepo.GetLowStock(ctx)
}

func (u *productUseCase) UpdateStock(ctx context.Context, id string, quantity int32) (*domain.Product, error) {
	if quantity < 0 {
		return nil, domain.ErrInsufficientStock
	}
	return u.productRepo.UpdateStock(ctx, id, quantity)
}

func (u *productUseCase) GetExpiredProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.productRepo.GetExpired(ctx)
}

func (u *productUseCase) GetProductsByCategory(ctx context.Context, categoryID string, page, limit int) ([]*domain.Product, int, error) {
	return u.productRepo.GetByCategory(ctx, categoryID, page, limit)
}

func (u *productUseCase) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	c.ID = uuid.NewString()
	c.CreatedAt = time.Now().UTC()
	if err := u.categoryRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (u *productUseCase) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	return u.categoryRepo.GetAll(ctx)
}
