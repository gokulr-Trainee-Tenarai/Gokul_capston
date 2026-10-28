package service

import (
	"context"
	"errors"
	"strings"

	"ecommerce-microservices/services/product/model"
	"ecommerce-microservices/services/product/repository"
)

type ProductService interface {
	Create(context.Context, *model.Product) error
	GetByID(context.Context, int64) (*model.Product, error)
	GetAll(context.Context) ([]model.Product, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

func (s *productService) Create(ctx context.Context, p *model.Product) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if p.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	return s.repo.Create(ctx, p)
}

func (s *productService) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	if id <= 0 {
		return nil, errors.New("product id must be greater than zero")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *productService) GetAll(ctx context.Context) ([]model.Product, error) {
	return s.repo.GetAll(ctx)
}
