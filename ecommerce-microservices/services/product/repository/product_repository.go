package repository

import (
	"context"
	"errors"

	"ecommerce-microservices/services/product/model"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(context.Context, *model.Product) error
	GetByID(context.Context, int64) (*model.Product, error)
	GetAll(context.Context) ([]model.Product, error)
}

type PostgresProductRepository struct {
	db *gorm.DB
}

func NewPostgresProductRepository(db *gorm.DB) ProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	var out []model.Product
	err := r.db.WithContext(ctx).Order("id ASC").Find(&out).Error
	return out, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}