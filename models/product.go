package models // <--- Notice the package name change

import (
	"errors"
	"time"
)

// Product is Public (Capital P) so other packages can see it.
type Product struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the data is safe to save.
// This is a "Method" attached to the Product struct (like a Class method in C#).
func (p *Product) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if p.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	if p.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}
