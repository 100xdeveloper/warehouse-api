package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"warehouse-api/database"
	"warehouse-api/internal/utils"
	"warehouse-api/models"

	"github.com/go-chi/chi/v5"
)

// Repository defines the interface for product data access
type ProductRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]models.Product, error)
	Create(ctx context.Context, p models.Product) (models.Product, error)
	GetByID(ctx context.Context, id int) (models.Product, error)
	Update(ctx context.Context, id int, p models.Product) error
	Delete(ctx context.Context, id int) error
}

type ProductHandler struct {
	Store ProductRepository
}

func NewProductHandler(store *database.ProductDB) *ProductHandler {
	return &ProductHandler{Store: store}
}

// ProductResponse defines the JSON structure for list requests
type ProductResponse struct {
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Data  []models.Product `json:"data"`
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	products, err := h.Store.GetAll(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	resp := ProductResponse{
		Page:  page,
		Limit: limit,
		Data:  products,
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newProduct, err := h.Store.Create(r.Context(), p)
	if err != nil {
		http.Error(w, "Failed to save product", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, newProduct)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	p, err := h.Store.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, http.StatusOK, p)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.Store.Update(r.Context(), id, p)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.Store.Delete(r.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
