package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"ecommerce-microservices/pkg/resilience"
	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/repository"
)

type ProductDTO struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type OrderService interface {
	Create(ctx context.Context, customerID int, productID int64, quantity int) (*model.Order, error)
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	ListAll(ctx context.Context) ([]*model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error)
	GetOrderWithProduct(ctx context.Context, orderID int64) (map[string]interface{}, error)
}

type orderService struct {
	repo              repository.OrderRepository
	productServiceURL string
	httpClient        *http.Client
	cb                *resilience.CircuitBreaker
}

func NewOrderService(repo repository.OrderRepository, productServiceURL string) OrderService {
	return &orderService{
		repo:              repo,
		productServiceURL: productServiceURL,
		httpClient:        &http.Client{Timeout: 3 * time.Second},
		cb:                resilience.NewCircuitBreaker(3, 10*time.Second),
	}
}

func (s *orderService) getProductDetails(ctx context.Context, productID int64) (*ProductDTO, error) {
	var prod ProductDTO

	op := func() error {
		// Pass incoming context into ExecuteWithRetry to handle context cancellation gracefully
		return resilience.ExecuteWithRetry(ctx, 3, 100*time.Millisecond, func() error {
			productURL := fmt.Sprintf("%s/products/%d", s.productServiceURL, productID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, productURL, nil)
			if err != nil {
				return err
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("network call failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("product %d not found", productID)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("product service returned status %d", resp.StatusCode)
			}

			return json.NewDecoder(resp.Body).Decode(&prod)
		})
	}

	if err := s.cb.Execute(op); err != nil {
		return nil, fmt.Errorf("failed to fetch product from product-service: %w", err)
	}

	return &prod, nil
}

func (s *orderService) Create(ctx context.Context, customerID int, productID int64, quantity int) (*model.Order, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	prod, err := s.getProductDetails(ctx, productID)
	if err != nil {
		return nil, err
	}

	order := &model.Order{
		CustomerID: customerID,
		ProductID:  productID,
		Quantity:   quantity,
		Total:      prod.Price * float64(quantity),
		Status:     model.StatusCreated,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListAll(ctx context.Context) ([]*model.Order, error) {
	return s.repo.ListAll(ctx)
}

func (s *orderService) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status == model.StatusCancelled {
		return nil, errors.New("cannot update status of a cancelled order")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *orderService) GetOrderWithProduct(ctx context.Context, orderID int64) (map[string]interface{}, error) {
	order, err := s.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	prod, err := s.getProductDetails(ctx, order.ProductID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"order":   order,
		"product": prod,
	}, nil
}
