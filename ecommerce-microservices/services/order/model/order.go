package model

import "time"

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID int         `gorm:"column:customer_id;not null;default:101" json:"customerId"`
	ProductID  int64       `gorm:"column:product_id;not null" json:"productId"`
	Quantity   int         `gorm:"not null" json:"quantity"`
	Total      float64     `gorm:"type:numeric(10,2);not null" json:"total"`
	Status     OrderStatus `gorm:"size:50;not null;default:'CREATED'" json:"status"`
	CreatedAt  time.Time   `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Order) TableName() string {
	return "orders"
}

type CreateOrderRequest struct {
	CustomerID int   `json:"customerId" binding:"required,gt=0"`
	ProductID  int64 `json:"productId" binding:"required,gt=0"`
	Quantity   int   `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required,oneof=CREATED CONFIRMED CANCELLED"`
}