package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"warehouse-api/models"

	"github.com/go-chi/chi/v5"
)

// MockStore is a mock implementation of ProductRepository
type MockStore struct {
	Products  []models.Product
	Err       error
	CreateFn  func(ctx context.Context, p models.Product) (models.Product, error)
	GetByIDFn func(ctx context.Context, id int) (models.Product, error)
	UpdateFn  func(ctx context.Context, id int, p models.Product) error
	DeleteFn  func(ctx context.Context, id int) error
}

func (m *MockStore) GetAll(ctx context.Context, limit, offset int) ([]models.Product, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	// Simple slice logic for mock
	start := offset
	if start >= len(m.Products) {
		return []models.Product{}, nil
	}
	end := start + limit
	if end > len(m.Products) {
		end = len(m.Products)
	}
	return m.Products[start:end], nil
}

func (m *MockStore) Create(ctx context.Context, p models.Product) (models.Product, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, p)
	}
	p.ID = len(m.Products) + 1
	p.CreatedAt = time.Now()
	m.Products = append(m.Products, p)
	return p, nil
}

func (m *MockStore) GetByID(ctx context.Context, id int) (models.Product, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	for _, p := range m.Products {
		if p.ID == id {
			return p, nil
		}
	}
	return models.Product{}, errors.New("product not found")
}

func (m *MockStore) Update(ctx context.Context, id int, p models.Product) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, p)
	}
	return nil
}

func (m *MockStore) Delete(ctx context.Context, id int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func TestGetProducts(t *testing.T) {
	mockStore := &MockStore{
		Products: []models.Product{
			{ID: 1, Name: "A", Price: 100, Stock: 10},
			{ID: 2, Name: "B", Price: 200, Stock: 20},
		},
	}
	handler := &ProductHandler{Store: mockStore}

	req := httptest.NewRequest("GET", "/products?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetProducts(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var data ProductResponse
	json.NewDecoder(resp.Body).Decode(&data)
	if len(data.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(data.Data))
	}
}

func TestCreateProduct_Success(t *testing.T) {
	mockStore := &MockStore{}
	handler := &ProductHandler{Store: mockStore}

	p := models.Product{Name: "New", Price: 100, Stock: 10}
	body, _ := json.Marshal(p)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestCreateProduct_ValidationFail(t *testing.T) {
	mockStore := &MockStore{}
	handler := &ProductHandler{Store: mockStore}

	p := models.Product{Name: "", Price: -10} // Invalid
	body, _ := json.Marshal(p)
	req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid input, got %d", w.Code)
	}
}

func TestGetProductByID_Found(t *testing.T) {
	mockStore := &MockStore{
		Products: []models.Product{
			{ID: 1, Name: "Exists", Price: 100},
		},
	}
	handler := &ProductHandler{Store: mockStore}

	req := httptest.NewRequest("GET", "/products/1", nil)
	// We need to inject the URL params because chi.URLParam uses context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetProductByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
