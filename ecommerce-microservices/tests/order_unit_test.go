package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/service"
)

type mockOrderRepo struct {
	createFunc       func(ctx context.Context, o *model.Order) error
	getByIDFunc      func(ctx context.Context, id int64) (*model.Order, error)
	listAllFunc      func(ctx context.Context) ([]*model.Order, error)
	updateStatusFunc func(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error)
}

func (m *mockOrderRepo) Create(ctx context.Context, o *model.Order) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, o)
	}
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderRepo) ListAll(ctx context.Context) ([]*model.Order, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil, nil
}

func TestCreateOrder_ZeroOrNegativeQuantity(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := service.NewOrderService(repo, "http://localhost:8081")

	_, err := svc.Create(context.Background(), 101, 1, 0)
	if err == nil {
		t.Fatalf("expected error for quantity <= 0, got nil")
	}
}

func TestUpdateStatus_CancelledOrder(t *testing.T) {
	repo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Order, error) {
			return &model.Order{
				ID:     id,
				Status: model.StatusCancelled,
			}, nil
		},
	}

	svc := service.NewOrderService(repo, "http://localhost:8081")

	_, err := svc.UpdateStatus(context.Background(), 1, model.StatusConfirmed)
	if err == nil {
		t.Fatalf("expected error when updating a cancelled order, got nil")
	}
}

func TestGetOrderByID_Success(t *testing.T) {
	expectedOrder := &model.Order{
		ID:         1,
		CustomerID: 101,
		ProductID:  10,
		Quantity:   2,
		Total:      199.98,
		Status:     model.StatusCreated,
		CreatedAt:  time.Now(),
	}

	repo := &mockOrderRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Order, error) {
			return expectedOrder, nil
		},
	}

	svc := service.NewOrderService(repo, "http://localhost:8081")

	result, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != expectedOrder.ID || result.Total != expectedOrder.Total {
		t.Errorf("expected order %+v, got %+v", expectedOrder, result)
	}
}

func TestConcurrentOrderRequests_100Goroutines(t *testing.T) {
	repo := &mockOrderRepo{
		createFunc: func(ctx context.Context, o *model.Order) error {
			return nil
		},
	}
	svc := service.NewOrderService(repo, "http://localhost:8081")

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(customerID int) {
			defer wg.Done()
			_, _ = svc.Create(context.Background(), customerID, 1, 2)
		}(100 + i)
	}

	wg.Wait()
}

func TestCreateOrder_ContextCancellation(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := service.NewOrderService(repo, "http://localhost:9999") // unreachable target

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	start := time.Now()
	_, err := svc.Create(ctx, 101, 1, 2)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}

	if duration > 100*time.Millisecond {
		t.Errorf("expected immediate exit under 100ms, took %v", duration)
	}
}
