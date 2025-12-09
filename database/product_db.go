package database

import (
	"context"
	"errors"
	"warehouse-api/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductDB holds the connection pool
type ProductDB struct {
	Pool *pgxpool.Pool
}

// New creates a new instance of ProductDB
func New(pool *pgxpool.Pool) *ProductDB {
	return &ProductDB{Pool: pool}
}

// --- METHODS (The SQL Logic) ---

// GetAll fetches products with pagination
// limit: how many to show
// offset: how many to skip
func (db *ProductDB) GetAll(ctx context.Context, limit, offset int) ([]models.Product, error) {
	// We add LIMIT and OFFSET to the SQL
	query := `SELECT id, name, price, stock, created_at
	          FROM products
	          ORDER BY id DESC
	          LIMIT $1 OFFSET $2`

	rows, err := db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if products == nil {
		products = []models.Product{}
	}

	return products, nil
}

func (db *ProductDB) Create(ctx context.Context, p models.Product) (models.Product, error) {
	query := `INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := db.Pool.QueryRow(ctx, query, p.Name, p.Price, p.Stock).Scan(&p.ID, &p.CreatedAt)
	return p, err
}

// GetByID fetches a single product
func (db *ProductDB) GetByID(ctx context.Context, id int) (models.Product, error) {
	var p models.Product
	query := `SELECT id, name, price, stock, created_at FROM products WHERE id = $1`

	err := db.Pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return p, errors.New("product not found")
		}
		return p, err
	}
	return p, nil
}

// Update modifies an existing product
func (db *ProductDB) Update(ctx context.Context, id int, p models.Product) error {
	query := `UPDATE products SET name=$1, price=$2, stock=$3 WHERE id=$4`

	commandTag, err := db.Pool.Exec(ctx, query, p.Name, p.Price, p.Stock, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("product not found")
	}
	return nil
}

// Delete removes a product
func (db *ProductDB) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`

	commandTag, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("product not found")
	}
	return nil
}
