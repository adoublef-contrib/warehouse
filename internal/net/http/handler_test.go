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

func Test_handleGetCategory(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		s, ctx := newServer(t)

		resp, err := post(ctx, s, "/categories", strings.NewReader(`{"name":"test"}`), contentType("application/json"))
		is.OK(t, err) // POST /categories
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cat := decode[struct{ ID int64 }](t, resp.Body)

		resp, err = get(ctx, s, "/categories/"+strconv.Itoa(int(cat.ID)))
		is.OK(t, err) // GET /categories/{category}
		is.Equal(t, resp.StatusCode, http.StatusOK)

		cc := decode[struct{ Data []category.Category }](t, resp.Body)
		is.Equal(t, len(cc.Data), 1)
		is.Equal(t, cc.Data[0].Name, "test")
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

	defer r.Close()

	var v V
	err := json.NewDecoder(r).Decode(&v)
	is.OK(t, err)
	return v
}

func contentType(s string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Add("Content-Type", s)
	}
}
