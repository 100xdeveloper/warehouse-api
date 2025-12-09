package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"warehouse-api/database"
	customMiddleware "warehouse-api/middleware"
	"warehouse-api/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Config holds the dependencies for our application.
// This replaces the global variables.
type Config struct {
	Store *database.ProductDB
}

// ProductResponse defines the JSON structure for list requests
type ProductResponse struct {
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Data  []models.Product `json:"data"`
}

func main() {
	// 1. Setup
	_ = godotenv.Load()
	dbUrl := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// 2. Initialize our "Repository"
	store := database.New(pool)

	// 3. Create the App Config (Inject the repository)
	app := Config{
		Store: store,
	}

	// 4. Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Note: We are calling methods on 'app', not global functions
	// A. Public Routes (No Key Needed)
	r.Group(func(r chi.Router) {
		r.Get("/products", app.getProducts)
		r.Get("/products/{id}", app.getProductByID)
	})

	// B. Protected Routes (Require API Key)
	r.Group(func(r chi.Router) {
		// Apply the middleware ONLY to this group
		r.Use(customMiddleware.APIKeyMiddleware)

		r.Post("/products", app.createProduct)
		r.Put("/products/{id}", app.updateProduct)
		r.Delete("/products/{id}", app.deleteProduct)
	})

	// 5. Start
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("Server running on port " + port)
	http.ListenAndServe(":"+port, r)
}

// Handlers
func (app *Config) getProducts(w http.ResponseWriter, r *http.Request) {
	// 1. Read Query Params (e.g. /products?page=2&limit=5)
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// 2. Set Defaults (if user didn't provide them)
	page := 1
	limit := 10

	// 3. Parse inputs (Convert string to int)
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 { // Max limit 100 for safety
			limit = l
		}
	}

	// 4. Calculate Offset
	// Page 1: skip 0. Page 2: skip 10. Page 3: skip 20.
	offset := (page - 1) * limit

	// 5. Call DB with new math
	products, err := app.Store.GetAll(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 6. Wrap the data in the struct
	resp := ProductResponse{
		Page:  page,
		Limit: limit,
		Data:  products,
	}

	// 7. Send the struct (resp) instead of the raw list (products)
	jsonResponse(w, http.StatusOK, resp)
}

func (app *Config) createProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validation logic is in the model
	if err := p.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// DB logic is in the database package
	newProduct, err := app.Store.Create(r.Context(), p)
	if err != nil {
		http.Error(w, "Failed to save product", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, newProduct)
}

func (app *Config) getProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	p, err := app.Store.GetByID(r.Context(), id)
	if err != nil {
		// Simple check: if error message contains "not found", return 404
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, http.StatusOK, p)
}

func (app *Config) updateProduct(w http.ResponseWriter, r *http.Request) {
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

	if err := p.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.Store.Update(r.Context(), id, p)
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

func (app *Config) deleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = app.Store.Delete(r.Context(), id)
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

// jsonResponse is a helper to send JSON status and data
// In Go, we hate repeating code. Let's create a Helper Function to handle sending JSON.
// This is very common in C# (like Ok(data)), and we can build it ourselves in Go.
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
