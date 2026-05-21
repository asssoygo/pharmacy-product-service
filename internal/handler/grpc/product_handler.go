package grpc

import (
	"context"
	"time"

	productpb "github.com/asssoygo/pharmacy-proto/gen/go/product"
	"github.com/asssoygo/pharmacy-product-service/internal/domain"
	"github.com/asssoygo/pharmacy-product-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductHandler struct {
	productpb.UnimplementedProductServiceServer
	uc usecase.ProductUseCase
}

func NewProductHandler(uc usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	p := &domain.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Quantity:      req.Quantity,
		CategoryID:    req.CategoryId,
		MinStockLevel: req.MinStockLevel,
	}
	if req.ExpiryDate != nil {
		t := req.ExpiryDate.AsTime()
		p.ExpiryDate = &t
	}
	created, err := h.uc.CreateProduct(ctx, p)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.CreateProductResponse{Product: domainToProto(created)}, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	p, err := h.uc.GetProduct(ctx, req.Id)
	if err != nil {
		if err == domain.ErrProductNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.GetProductResponse{Product: domainToProto(p)}, nil
}

func (h *ProductHandler) GetProducts(ctx context.Context, req *productpb.GetProductsRequest) (*productpb.GetProductsResponse, error) {
	page, limit := int(req.Page), int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	products, total, err := h.uc.GetProducts(ctx, page, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.GetProductsResponse{Products: domainSliceToProto(products), Total: int32(total)}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	p := &domain.Product{
		ID:            req.Id,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		CategoryID:    req.CategoryId,
		MinStockLevel: req.MinStockLevel,
	}
	if req.ExpiryDate != nil {
		t := req.ExpiryDate.AsTime()
		p.ExpiryDate = &t
	}
	updated, err := h.uc.UpdateProduct(ctx, p)
	if err != nil {
		if err == domain.ErrProductNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.UpdateProductResponse{Product: domainToProto(updated)}, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	if err := h.uc.DeleteProduct(ctx, req.Id); err != nil {
		if err == domain.ErrProductNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.DeleteProductResponse{Success: true}, nil
}

func (h *ProductHandler) SearchProducts(ctx context.Context, req *productpb.SearchProductsRequest) (*productpb.SearchProductsResponse, error) {
	page, limit := int(req.Page), int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	products, total, err := h.uc.SearchProducts(ctx, req.Query, page, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.SearchProductsResponse{Products: domainSliceToProto(products), Total: int32(total)}, nil
}

func (h *ProductHandler) GetLowStockProducts(ctx context.Context, _ *productpb.GetLowStockProductsRequest) (*productpb.GetLowStockProductsResponse, error) {
	products, err := h.uc.GetLowStockProducts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.GetLowStockProductsResponse{Products: domainSliceToProto(products)}, nil
}

func (h *ProductHandler) UpdateStock(ctx context.Context, req *productpb.UpdateStockRequest) (*productpb.UpdateStockResponse, error) {
	p, err := h.uc.UpdateStock(ctx, req.Id, req.Quantity)
	if err != nil {
		if err == domain.ErrProductNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if err == domain.ErrInsufficientStock {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.UpdateStockResponse{Product: domainToProto(p)}, nil
}

func (h *ProductHandler) GetExpiredProducts(ctx context.Context, _ *productpb.GetExpiredProductsRequest) (*productpb.GetExpiredProductsResponse, error) {
	products, err := h.uc.GetExpiredProducts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.GetExpiredProductsResponse{Products: domainSliceToProto(products)}, nil
}

func (h *ProductHandler) GetProductsByCategory(ctx context.Context, req *productpb.GetProductsByCategoryRequest) (*productpb.GetProductsByCategoryResponse, error) {
	page, limit := int(req.Page), int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	products, total, err := h.uc.GetProductsByCategory(ctx, req.CategoryId, page, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.GetProductsByCategoryResponse{Products: domainSliceToProto(products), Total: int32(total)}, nil
}

func (h *ProductHandler) CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (*productpb.CreateCategoryResponse, error) {
	c := &domain.Category{Name: req.Name, Description: req.Description}
	created, err := h.uc.CreateCategory(ctx, c)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &productpb.CreateCategoryResponse{Category: categoryToProto(created)}, nil
}

func (h *ProductHandler) GetCategories(ctx context.Context, _ *productpb.GetCategoriesRequest) (*productpb.GetCategoriesResponse, error) {
	cats, err := h.uc.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var out []*productpb.Category
	for _, c := range cats {
		out = append(out, categoryToProto(c))
	}
	return &productpb.GetCategoriesResponse{Categories: out}, nil
}

func domainToProto(p *domain.Product) *productpb.Product {
	pb := &productpb.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		Quantity:      p.Quantity,
		CategoryId:    p.CategoryID,
		MinStockLevel: p.MinStockLevel,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
	if p.ExpiryDate != nil {
		pb.ExpiryDate = timestamppb.New(*p.ExpiryDate)
	}
	return pb
}

func domainSliceToProto(products []*domain.Product) []*productpb.Product {
	out := make([]*productpb.Product, len(products))
	for i, p := range products {
		out[i] = domainToProto(p)
	}
	return out
}

func categoryToProto(c *domain.Category) *productpb.Category {
	return &productpb.Category{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(time.Time{}),
	}
}
