package category_test

import (
	"context"
	"testing"

	. "github.com/adoublef-contrib/warehouse-management/internal/category"
	"github.com/adoublef-contrib/warehouse-management/internal/database/postgres"
	"go.adoublef.dev/testing/is"
)

func TestDB_AddCategory(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := newDB(t)
		ctx := t.Context()

		var (
			name = "name"
		)
		id, err := db.AddCategory(ctx, name)
		is.OK(t, err) // DB.AddCategory
		is.Equal(t, id, 1)
	})
}

func TestDB_UpdateCategory(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := newDB(t)
		ctx := t.Context()

		var (
			name = "name"
		)
		id, err := db.AddCategory(ctx, name)
		is.OK(t, err) // DB.AddCategory
		is.Equal(t, id, 1)

		err = db.UpdateCategory(ctx, "new", id)
		is.OK(t, err) // DB.UpdateCategory

		c, err := db.Category(ctx, id)
		is.OK(t, err) // DB.Category

		is.Equal(t, c.Name, "new")
	})
}

func TestDB_DeleteCategory(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := newDB(t)
		ctx := t.Context()

		var (
			name = "name"
		)
		id, err := db.AddCategory(ctx, name)
		is.OK(t, err) // DB.AddCategory
		is.Equal(t, id, 1)

		_, err = db.DeleteCategory(ctx, id)
		is.OK(t, err) // DB.DeleteCategory
	})
}

func TestDB_Product(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := newDB(t)
		ctx := t.Context()

		var (
			name = "cat"
		)
		cid, err := db.AddCategory(ctx, name)
		is.OK(t, err) // DB.AddCategory
		is.Equal(t, cid, 1)

		pid, err := db.AddProduct(ctx, "product", 1, "cat")
		is.OK(t, err) // DB.AddProduct

		p, err := db.Product(ctx, pid)
		is.OK(t, err) // DB.Product

		is.Equal(t, p.Stock, 1)
	})
}

func TestDB_DeleteProduct(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		db := newDB(t)
		ctx := t.Context()

		var (
			name = "cat"
		)
		cid, err := db.AddCategory(ctx, name)
		is.OK(t, err) // DB.AddCategory
		is.Equal(t, cid, 1)

		pid, err := db.AddProduct(ctx, "product", 1, "cat")
		is.OK(t, err) // DB.AddProduct

		err = db.DeleteProduct(ctx, pid)
		is.OK(t, err) // DB.DeleteProduct
	})
}

func newDB(t testing.TB) *DB {
	t.Helper()

	tCtx := t.Context()

	p, err := container.ConnectionPool(tCtx)
	is.OK(t, err) // container.ConnectionPool(ctx)
	t.Cleanup(func() { p.Close() })

	fsys := &postgres.FS{
		URL: p.Config().ConnString(),
	}
	is.OK(t, fsys.Up(tCtx)) // fsys.Up
	t.Cleanup(func() { fsys.Down(context.Background()) })

	return &DB{RWC: p}
}
