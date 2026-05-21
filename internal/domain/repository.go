package domain

import "context"

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, page, limit int) ([]*Product, int, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	SearchByNameOrCategory(ctx context.Context, query string, page, limit int) ([]*Product, int, error)
	GetLowStock(ctx context.Context) ([]*Product, error)
	GetExpired(ctx context.Context) ([]*Product, error)
	GetByCategory(ctx context.Context, categoryID string, page, limit int) ([]*Product, int, error)
	UpdateStock(ctx context.Context, id string, quantity int32) (*Product, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	GetAll(ctx context.Context) ([]*Category, error)
	GetByID(ctx context.Context, id string) (*Category, error)
}
