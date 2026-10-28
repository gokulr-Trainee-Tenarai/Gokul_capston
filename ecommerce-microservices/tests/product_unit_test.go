package tests

import (
	"context"
	"sync"
	"testing"

	"ecommerce-microservices/services/product/model"
	"ecommerce-microservices/services/product/service"
)

type mockProductRepo struct {
	createFunc  func(ctx context.Context, p *model.Product) error
	getByIDFunc func(ctx context.Context, id int64) (*model.Product, error)
	getAllFunc  func(ctx context.Context) ([]model.Product, error)
}

func (m *mockProductRepo) Create(ctx context.Context, p *model.Product) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return nil
}

func (m *mockProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockProductRepo) GetAll(ctx context.Context) ([]model.Product, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx)
	}
	return nil, nil
}

func TestProductCreate_EmptyName(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	err := svc.Create(context.Background(), &model.Product{
		Name:  "   ",
		Price: 49.99,
	})

	if err == nil {
		t.Fatalf("expected error for empty name, got nil")
	}
}

func TestProductCreate_InvalidPrice(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	err := svc.Create(context.Background(), &model.Product{
		Name:  "Mechanical Keyboard",
		Price: -10.0,
	})

	if err == nil {
		t.Fatalf("expected error for price <= 0, got nil")
	}
}

func TestProductCreate_Success(t *testing.T) {
	repo := &mockProductRepo{
		createFunc: func(ctx context.Context, p *model.Product) error {
			p.ID = 100
			return nil
		},
	}
	svc := service.NewProductService(repo)

	prod := &model.Product{
		Name:  "Mechanical Keyboard",
		Price: 89.99,
	}

	err := svc.Create(context.Background(), prod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if prod.ID != 100 {
		t.Errorf("expected product ID to be 100, got %d", prod.ID)
	}
}

func TestProduct_GetByID_InvalidID(t *testing.T) {
	repo := &mockProductRepo{}
	svc := service.NewProductService(repo)

	_, err := svc.GetByID(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for ID <= 0, got nil")
	}

	_, err = svc.GetByID(context.Background(), -5)
	if err == nil {
		t.Fatal("expected error for negative ID, got nil")
	}
}

func TestProduct_ConcurrentReads_100Goroutines(t *testing.T) {
	repo := &mockProductRepo{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Product, error) {
			return &model.Product{ID: id, Name: "Monitor", Price: 329.00}, nil
		},
	}
	svc := service.NewProductService(repo)

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int64) {
			defer wg.Done()
			_, err := svc.GetByID(context.Background(), id)
			if err != nil {
				t.Errorf("concurrent read error: %v", err)
			}
		}(int64(i + 1))
	}

	wg.Wait()
}
