package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asssoygo/pharmacy-product-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	productTTL  = 5 * time.Minute
	productsTTL = 5 * time.Minute
)

type CachedProductRepository struct {
	repo  domain.ProductRepository
	redis *redis.Client
}

func NewCachedProductRepository(repo domain.ProductRepository, client *redis.Client) domain.ProductRepository {
	return &CachedProductRepository{repo: repo, redis: client}
}

func productKey(id string) string { return fmt.Sprintf("product:%s", id) }
func productsKey(page, limit int) string {
	return fmt.Sprintf("products:page:%d:limit:%d", page, limit)
}

func (r *CachedProductRepository) Create(ctx context.Context, p *domain.Product) error {
	return r.repo.Create(ctx, p)
}

func (r *CachedProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	key := productKey(id)
	data, err := r.redis.Get(ctx, key).Bytes()
	if err == nil {
		var p domain.Product
		if json.Unmarshal(data, &p) == nil {
			return &p, nil
		}
	}
	p, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b, e := json.Marshal(p); e == nil {
		r.redis.Set(ctx, key, b, productTTL)
	}
	return p, nil
}

func (r *CachedProductRepository) List(ctx context.Context, page, limit int) ([]*domain.Product, int, error) {
	key := productsKey(page, limit)
	type cached struct {
		Products []*domain.Product
		Total    int
	}
	data, err := r.redis.Get(ctx, key).Bytes()
	if err == nil {
		var c cached
		if json.Unmarshal(data, &c) == nil {
			return c.Products, c.Total, nil
		}
	}
	products, total, err := r.repo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	if b, e := json.Marshal(cached{Products: products, Total: total}); e == nil {
		r.redis.Set(ctx, key, b, productsTTL)
	}
	return products, total, nil
}

func (r *CachedProductRepository) Update(ctx context.Context, p *domain.Product) error {
	if err := r.repo.Update(ctx, p); err != nil {
		return err
	}
	r.invalidate(ctx, p.ID)
	return nil
}

func (r *CachedProductRepository) Delete(ctx context.Context, id string) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, id)
	return nil
}

func (r *CachedProductRepository) UpdateStock(ctx context.Context, id string, quantity int32) (*domain.Product, error) {
	p, err := r.repo.UpdateStock(ctx, id, quantity)
	if err != nil {
		return nil, err
	}
	r.invalidate(ctx, id)
	return p, nil
}

func (r *CachedProductRepository) invalidate(ctx context.Context, id string) {
	r.redis.Del(ctx, productKey(id))
	keys, _ := r.redis.Keys(ctx, "products:page:*").Result()
	if len(keys) > 0 {
		r.redis.Del(ctx, keys...)
	}
}

func (r *CachedProductRepository) SearchByNameOrCategory(ctx context.Context, query string, page, limit int) ([]*domain.Product, int, error) {
	return r.repo.SearchByNameOrCategory(ctx, query, page, limit)
}

func (r *CachedProductRepository) GetLowStock(ctx context.Context) ([]*domain.Product, error) {
	return r.repo.GetLowStock(ctx)
}

func (r *CachedProductRepository) GetExpired(ctx context.Context) ([]*domain.Product, error) {
	return r.repo.GetExpired(ctx)
}

func (r *CachedProductRepository) GetByCategory(ctx context.Context, categoryID string, page, limit int) ([]*domain.Product, int, error) {
	return r.repo.GetByCategory(ctx, categoryID, page, limit)
}

