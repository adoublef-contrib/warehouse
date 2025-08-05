package category

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	RWC *pgxpool.Pool
}

func (d *DB) SearchCategory(ctx context.Context, name string) (*Category, error) {
	const stmt = "SELECT category_id, category_name FROM categories WHERE category_name = $1"
	r := d.RWC.QueryRow(ctx, stmt, name)
	var c Category
	err := r.Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, fmt.Errorf("Error querying category: %v", err)
	}
	return &c, nil
}

func (d *DB) Category(ctx context.Context, id int64) (*Category, error) {
	const stmt = "SELECT category_id, category_name FROM categories WHERE category_id = $1"
	r := d.RWC.QueryRow(ctx, stmt, id)
	var c Category
	err := r.Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, fmt.Errorf("Error querying category: %v", err)
	}
	return &c, nil
}

func (d *DB) AddCategory(ctx context.Context, name string) (int64, error) {
	var id int64
	err := d.RWC.QueryRow(ctx, "INSERT INTO categories (category_name) VALUES ($1) RETURNING category_id", name).
		Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("Error inserting new row: %v", err)
	}
	return id, nil
}

func (d *DB) UpdateCategory(ctx context.Context, name string, id int64) (int64, error) {
	const stmt = "UPDATE categories SET category_name=$1 WHERE category_id=$2 RETURNING category_id"
	var uid int64
	err := d.RWC.QueryRow(ctx, stmt, name, id).
		Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("Category with ID %d not found: %w", id, err)
		} else {
			return 0, fmt.Errorf("Error querying category: %w", err)
		}
	}
	return uid, nil
}

func (d *DB) DeleteCategory(ctx context.Context, id int64) (int64, error) {
	const stmt = "DELETE FROM categories WHERE category_id = $1 RETURNING category_id"
	var did int64
	err := d.RWC.QueryRow(ctx, stmt, id).Scan(&did)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("Category with ID %d not found: %w", did, err)
		} else {
			return 0, fmt.Errorf("Error querying category: %w", err)
		}
	}
	return did, nil
}

func (d *DB) Product(ctx context.Context, id int64) (*Product, error) {
	const stmt = "SELECT * FROM products WHERE product_id=$1"
	row := d.RWC.QueryRow(ctx, stmt, id)
	var p Product
	err := row.Scan(&p.ID, &p.Category, &p.Name, &p.Stock)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product with id %d not found", id)
		}
		return nil, err
	}
	return &p, nil
}

func (d *DB) SearchProduct(ctx context.Context, name string) ([]*Product, error) {
	const stmt = "SELECT product_id, product_name, stock FROM products WHERE product_name ILIKE $1"
	query := "%" + name + "%"
	rows, err := d.RWC.Query(ctx, stmt, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var productList []*Product
	for rows.Next() {
		p := &Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Stock); err != nil {
			return nil, err
		}
		productList = append(productList, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return productList, nil
}

func (d *DB) ProductsByCategory(ctx context.Context, category string) ([]Product, error) {
	const stmt = "SELECT p.product_id, p.product_name, p.stock, c.category_name FROM products as p INNER JOIN categories as c ON p.category_id = c.category_id WHERE c.category_name = $1;"
	rr, err := d.RWC.Query(ctx, stmt, category)
	if err != nil {
		return nil, err
	}
	defer rr.Close()
	var pp []Product
	for rr.Next() {
		var p Product
		if err := rr.Scan(&p.ID, &p.Name, &p.Stock, &p.Category); err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	if err := rr.Err(); err != nil {
		return nil, err
	}
	return pp, nil
}

func (d *DB) Products(ctx context.Context, limit int, offset int) ([]*Product, error) {
	const stmt = "SELECT product_id, product_name, stock FROM products ORDER BY product_id LIMIT $1 OFFSET $2"
	rows, err := d.RWC.Query(ctx, stmt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var productList []*Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Stock); err != nil {
			return nil, err
		}
		productList = append(productList, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return productList, nil
}

func (d *DB) AddProduct(ctx context.Context, name string, stock int, category string) (int64, error) {
	// Must be modified to not allow duplicate entries
	const stmt = "INSERT INTO products (product_name, stock, category_id) VALUES ($1, $2, (SELECT category_id FROM categories WHERE category_name = $3)) RETURNING product_id"
	var id int64
	err := d.RWC.QueryRow(ctx, stmt, name, stock, category).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert product: %w", err)
	}
	return id, nil
}

func (d *DB) UpdateProduct(ctx context.Context, id int64, name string, stock int, category string) error {
	const stmt = "UPDATE products SET product_name = $1, stock = $2, category_id = (SELECT category_id from categories WHERE category_name = $3) WHERE product_id = $4"
	result, err := d.RWC.Exec(ctx, stmt, name, stock, category, id)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("failed to file product found with ID %d", id)
	}
	return nil
}

func (d *DB) DeleteProduct(ctx context.Context, id int64) error {
	const stmt = "DELETE FROM products WHERE product_id = $1"
	result, err := d.RWC.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("failed to file product found with ID %d", id)
	}
	return nil
}
