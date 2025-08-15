package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/adoublef-contrib/warehouse-management/internal/category"
	"github.com/adoublef-contrib/warehouse-management/internal/database/postgres"
	. "github.com/adoublef-contrib/warehouse-management/internal/net/http"
	"go.adoublef.dev/testing/is"
)

func Test_handleAddNewProduct(t *testing.T) {
	t.Run("ErrCategory", func(t *testing.T) {
		s, ctx := newServer(t)

		resp, err := post(ctx, s, "/products", strings.NewReader(`{"name":"product","stock":10,"category":"never"}`), contentType("application/json"))
		is.OK(t, err) // POST /products
		is.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func Test_handleGetProduct(t *testing.T) {
	s, ctx := newServer(t)

	// require a category, still worth checking errors.
	resp, err := post(ctx, s, "/categories", strings.NewReader(`{"name":"category"}`), contentType("application/json"))
	is.OK(t, err) // POST /categories
	is.Equal(t, resp.StatusCode, http.StatusOK)
	close(t, resp.Body)

	resp, err = post(ctx, s, "/products", strings.NewReader(`{"name":"product","stock":10,"category":"category"}`), contentType("application/json"))
	is.OK(t, err) // POST /products

	product := decode[struct{ ID int64 }](t, resp.Body)

	t.Run("OK", func(t *testing.T) {
		resp, err = get(ctx, s, "/products/"+strconv.Itoa(int(product.ID)))
		is.OK(t, err) // GET /categories/{category}
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "product")
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		resp, err := get(ctx, s, "/products/2")
		is.OK(t, err) // GET /products/{product}
		is.Equal(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("SearchName", func(t *testing.T) {
		resp, err = get(ctx, s, "/products/search/name?product=product")
		is.OK(t, err) // GET /products/search/name
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "product")
	})

	t.Run("SearchCategory", func(t *testing.T) {
		resp, err = get(ctx, s, "/products/search/name?category=category")
		is.OK(t, err) // GET /products/search/name
		is.Equal(t, resp.StatusCode, http.StatusOK)
		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "product")
	})
}

func Test_handleGetCategory(t *testing.T) {
	s, ctx := newServer(t)

	resp, err := post(ctx, s, "/categories", strings.NewReader(`{"name":"test"}`), contentType("application/json"))
	is.OK(t, err) // POST /categories
	is.Equal(t, resp.StatusCode, http.StatusOK)

	cat := decode[struct{ ID int64 }](t, resp.Body)

	t.Run("OK", func(t *testing.T) {
		resp, err = get(ctx, s, "/categories/"+strconv.Itoa(int(cat.ID)))
		is.OK(t, err) // GET /categories/{category}
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "test")
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		resp, err := get(ctx, s, "/categories/2")
		is.OK(t, err) // GET /categories/{category}
		is.Equal(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("Search", func(t *testing.T) {
		resp, err = get(ctx, s, "/categories/search/name?category=test")
		is.OK(t, err) // GET /categories/search/name
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "test")
	})

	t.Run("ErrSearch", func(t *testing.T) {
		resp, err = get(ctx, s, "/categories/search/name?category=never")
		is.OK(t, err) // GET /categories/search/name
		is.Equal(t, resp.StatusCode, http.StatusNotFound)
	})
}

func newServer(t testing.TB) (*httptest.Server, context.Context) {
	t.Helper()

	ctx := t.Context()

	p, err := container.ConnectionPool(ctx)
	is.OK(t, err) // container.ConnectionPool(ctx)
	t.Cleanup(func() { p.Close() })

	fsys := &postgres.FS{
		URL: p.Config().ConnString(),
	}
	is.OK(t, fsys.Up(ctx)) // fsys.Up
	t.Cleanup(func() { fsys.Down(context.Background()) })

	h := Handler(&category.DB{RWC: p})
	s := httptest.NewServer(h)

	t.Cleanup(func() { s.Close() })
	return s, ctx
}

func contentType(s string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Add("Content-Type", s)
	}
}

func get(ctx context.Context, s *httptest.Server, path string, opts ...func(*http.Request)) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+path, nil)
	if err != nil {
		return nil, err
	}
	for _, f := range opts {
		f(req)
	}
	return s.Client().Do(req)
}

func post(ctx context.Context, s *httptest.Server, path string, body io.Reader, opts ...func(*http.Request)) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+path, body)
	if err != nil {
		return nil, err
	}
	for _, f := range opts {
		f(req)
	}
	return s.Client().Do(req)
}

func decode[V any](t testing.TB, r io.ReadCloser) V {
	t.Helper()

	var v V
	err := json.NewDecoder(r).Decode(&v)
	is.OK(t, err)
	close(t, r)

	return v
}

func close(t testing.TB, closer io.Closer) {
	t.Helper()
	is.OK(t, closer.Close())
}
