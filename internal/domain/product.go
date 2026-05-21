package domain

import (
	"errors"
	"time"
)

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

type Product struct {
	ID            string
	Name          string
	Description   string
	Price         float64
	Quantity      int32
	CategoryID    string
	CategoryName  string
	MinStockLevel int32
	ExpiryDate    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Category struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}
