package repository

import (
	"context"
	"errors"

	"ecommerce-microservices/services/order/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(context.Context, *model.Order) error
	GetByID(context.Context, int64) (*model.Order, error)
	ListAll(context.Context) ([]*model.Order, error)
	UpdateStatus(context.Context, int64, model.OrderStatus) (*model.Order, error)
}

type PostgresOrderRepository struct {
	db *gorm.DB
}

func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *model.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *PostgresOrderRepository) ListAll(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order
	err := r.db.WithContext(ctx).Order("id ASC").Find(&orders).Error
	return orders, err
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	o.Status = status
	if err := r.db.WithContext(ctx).Save(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
