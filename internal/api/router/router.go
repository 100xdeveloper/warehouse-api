package router

import (
	"warehouse-api/internal/api/handlers"
	customMiddleware "warehouse-api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(productHandler *handlers.ProductHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/products", productHandler.GetProducts)
		r.Get("/products/{id}", productHandler.GetProductByID)
	})

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.APIKeyMiddleware)
		r.Post("/products", productHandler.CreateProduct)
		r.Put("/products/{id}", productHandler.UpdateProduct)
		r.Delete("/products/{id}", productHandler.DeleteProduct)
	})

	return r
}
