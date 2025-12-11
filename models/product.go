package models // <--- Notice the package name change

import (
	"time"
)

// Product is Public (Capital P) so other packages can see it.
type Product struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required"`
	Price     int       `json:"price" validate:"required,gt=0"`
	Stock     int       `json:"stock" validate:"gte=0"`
	CreatedAt time.Time `json:"created_at"`
}
