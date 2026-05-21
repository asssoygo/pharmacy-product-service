package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/asssoygo/pharmacy-product-service/internal/domain"
	_ "github.com/lib/pq"
)

type productRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, p *domain.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, quantity, category_id, min_stock_level, expiry_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Quantity,
		p.CategoryID, p.MinStockLevel, p.ExpiryDate, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *productRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanProduct(row)
}

func (r *productRepo) List(ctx context.Context, page, limit int) ([]*domain.Product, int, error) {
	offset := (page - 1) * limit
	query := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	products, err := scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&total)
	return products, total, nil
}

func (r *productRepo) Update(ctx context.Context, p *domain.Product) error {
	query := `
		UPDATE products
		SET name=$2, description=$3, price=$4, quantity=$5, category_id=$6,
		    min_stock_level=$7, expiry_date=$8, updated_at=$9
		WHERE id=$1`
	res, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Quantity,
		p.CategoryID, p.MinStockLevel, p.ExpiryDate, p.UpdatedAt,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *productRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *productRepo) SearchByNameOrCategory(ctx context.Context, query string, page, limit int) ([]*domain.Product, int, error) {
	offset := (page - 1) * limit
	q := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.name ILIKE $1 OR c.name ILIKE $1
		ORDER BY p.name
		LIMIT $2 OFFSET $3`
	like := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, like, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	products, err := scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.name ILIKE $1 OR c.name ILIKE $1`,
		like,
	).Scan(&total)
	return products, total, nil
}

func (r *productRepo) GetLowStock(ctx context.Context) ([]*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.quantity <= p.min_stock_level`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *productRepo) GetExpired(ctx context.Context) ([]*domain.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.expiry_date IS NOT NULL AND p.expiry_date < NOW()`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *productRepo) GetByCategory(ctx context.Context, categoryID string, page, limit int) ([]*domain.Product, int, error) {
	offset := (page - 1) * limit
	query := `
		SELECT p.id, p.name, p.description, p.price, p.quantity, p.category_id,
		       COALESCE(c.name,'') AS category_name,
		       p.min_stock_level, p.expiry_date, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.category_id = $1
		ORDER BY p.name
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	products, err := scanProducts(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE category_id=$1`, categoryID).Scan(&total)
	return products, total, nil
}

func (r *productRepo) UpdateStock(ctx context.Context, id string, quantity int32) (*domain.Product, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET quantity=$2, updated_at=$3 WHERE id=$1`,
		id, quantity, time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func scanProduct(row *sql.Row) (*domain.Product, error) {
	p := &domain.Product{}
	var expiryDate sql.NullTime
	err := row.Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity,
		&p.CategoryID, &p.CategoryName, &p.MinStockLevel,
		&expiryDate, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	if expiryDate.Valid {
		p.ExpiryDate = &expiryDate.Time
	}
	return p, nil
}

func scanProducts(rows *sql.Rows) ([]*domain.Product, error) {
	var products []*domain.Product
	for rows.Next() {
		p := &domain.Product{}
		var expiryDate sql.NullTime
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity,
			&p.CategoryID, &p.CategoryName, &p.MinStockLevel,
			&expiryDate, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if expiryDate.Valid {
			p.ExpiryDate = &expiryDate.Time
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

type categoryRepo struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) domain.CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, c *domain.Category) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (id, name, description, created_at) VALUES ($1,$2,$3,$4)`,
		c.ID, c.Name, c.Description, c.CreatedAt,
	)
	return err
}

func (r *categoryRepo) GetAll(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []*domain.Category
	for rows.Next() {
		c := &domain.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *categoryRepo) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	c := &domain.Category{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at FROM categories WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCategoryNotFound
	}
	return c, err
}
