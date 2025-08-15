package http

import (
	"cmp"
	"net/http"
	"strconv"

	"github.com/adoublef-contrib/warehouse-management/internal/category"
)

type Server = http.Server

func Handler(db *category.DB) http.Handler {
	mux := http.NewServeMux()
	handleFunc := func(pattern string, h http.Handler) {
		mux.Handle(pattern, h)
	}

	handleFunc("GET /products/{product}", handleGetProductByID(db))
	handleFunc("GET /products/search/name", handleGetProductByName(db))
	handleFunc("GET /products", handleGetAllProducts(db))
	handleFunc("GET /products/search/category", handleGetProductsByCategory(db))
	handleFunc("POST /products", handleAddNewProduct(db))
	handleFunc("PATCH /products/{product}", handleUpdateProduct(db))
	handleFunc("DELETE /products/{product}", handleDeleteProduct(db))
	handleFunc("GET /categories/search/name", handleGetCategoryByName(db))
	handleFunc("GET /categories/{category}", handleGetCategoryByID(db))
	handleFunc("POST /categories", handleAddNewCategory(db))
	handleFunc("PATCH /categories/{category}", handleUpdateCategory(db))

	return mux
}

func handleGetProductByID(db *category.DB) http.HandlerFunc {
	type response struct {
		Data []category.Product `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (int64, error) {
		id, err := strconv.ParseInt(r.PathValue("product"), 10, 64)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		v, err := db.Product(ctx, id)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{Data: []category.Product{v}}, http.StatusOK)
	}
}

func handleGetProductByName(db *category.DB) http.HandlerFunc {
	type response struct {
		Data []category.Product `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (string, error) {
		return r.URL.Query().Get("name"), nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		v, err := db.SearchProduct(ctx, name)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{Data: v}, http.StatusOK)
	}
}

func handleGetAllProducts(db *category.DB) http.HandlerFunc {
	type response struct {
		Data []category.Product `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (lim, off int64, err error) {
		lim, err1 := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		off, err2 := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
		return lim, off, cmp.Or(err1, err2)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		lim, off, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		pp, err := db.Products(ctx, lim, off)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{pp}, http.StatusOK)
	}
}

func handleGetProductsByCategory(db *category.DB) http.HandlerFunc {
	type response struct {
		Data []category.Product `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (string, error) {
		return r.URL.Query().Get("category"), nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		category, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		v, err := db.ProductsByCategory(ctx, category)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{Data: v}, http.StatusOK)
	}
}

func handleAddNewProduct(db *category.DB) http.HandlerFunc {
	type request struct {
		Name     *string `json:"name"`
		Stock    int64   `json:"stock"`
		Category string  `json:"category"`
	}
	type response struct {
		ID int64 `json:"id"`
	}
	parse := func(w http.ResponseWriter, r *http.Request) (stock int64, name, category string, err error) {
		v, err := Decode[request](w, r, 0, 0)
		if err != nil {
			return 0, "", "", err
		}
		if v.Name == nil {
			return 0, "", "", newErr(http.StatusBadRequest, "name is required")
		}
		return v.Stock, *v.Name, v.Category, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		stock, name, category, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		id, err := db.AddProduct(ctx, name, stock, category)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{id}, http.StatusOK)
	}
}

func handleUpdateProduct(db *category.DB) http.HandlerFunc {
	type request struct {
		Name     *string `json:"name"`
		Stock    int64   `json:"stock"`
		Category string  `json:"string"`
	}
	parse := func(w http.ResponseWriter, r *http.Request) (id, stock int64, name, category string, err error) {
		id, err1 := strconv.ParseInt(r.PathValue("product"), 10, 64)
		v, err2 := Decode[request](w, r, 0, 0)
		if err := cmp.Or(err1, err2); err != nil {
			return 0, 0, "", "", err
		}
		if v.Name == nil {
			return 0, 0, "", "", newErr(http.StatusBadRequest, "name is required")
		}
		return id, v.Stock, *v.Name, v.Category, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, stock, name, category, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		if err := db.UpdateProduct(ctx, id, name, stock, category); err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent) // ok
	}
}

func handleDeleteProduct(db *category.DB) http.HandlerFunc {
	parse := func(_ http.ResponseWriter, r *http.Request) (int64, error) {
		id, err := strconv.ParseInt(r.PathValue("product"), 10, 64)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		if err := db.DeleteProduct(ctx, id); err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent) // ok
	}
}

func handleGetCategoryByName(db *category.DB) http.HandlerFunc {
	type request struct {
		Name *string `json:"name"`
	}
	type response struct {
		Data []category.Category `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (string, error) {
		return r.URL.Query().Get("category"), nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		c, err := db.SearchCategory(ctx, name)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{Data: []category.Category{c}}, http.StatusOK)
	}
}

func handleGetCategoryByID(db *category.DB) http.HandlerFunc {
	type request struct {
		Name *string `json:"name"`
	}
	type response struct {
		Data []category.Category `json:"data"`
	}
	parse := func(_ http.ResponseWriter, r *http.Request) (int64, error) {
		id, err := strconv.ParseInt(r.PathValue("category"), 10, 64)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		c, err := db.Category(ctx, id)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{Data: []category.Category{c}}, http.StatusOK)
	}
}

func handleAddNewCategory(db *category.DB) http.HandlerFunc {
	type request struct {
		Name *string `json:"name"`
	}
	type response struct {
		ID int64 `json:"id"`
	}
	parse := func(w http.ResponseWriter, r *http.Request) (string, error) {
		v, err := Decode[request](w, r, 0, 0)
		if err != nil {
			return "", err
		}
		if v.Name == nil {
			return "", newErr(http.StatusBadRequest, "name is required")
		}
		return *v.Name, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		id, err := db.AddCategory(ctx, name)
		if err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		respond(w, r, response{id}, http.StatusOK)
	}
}

func handleUpdateCategory(db *category.DB) http.HandlerFunc {
	type request struct {
		Name *string `json:"name"`
	}
	parse := func(w http.ResponseWriter, r *http.Request) (int64, string, error) {
		id, err1 := strconv.ParseInt(r.PathValue("category"), 10, 64)
		v, err2 := Decode[request](w, r, 0, 0)
		if err := cmp.Or(err1, err2); err != nil {
			return 0, "", err
		}
		if v.Name == nil {
			return 0, "", newErr(http.StatusBadRequest, "name is required")
		}
		return id, *v.Name, nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, name, err := parse(w, r)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		if err := db.UpdateCategory(ctx, name, id); err != nil {
			handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent) // ok
	}
}

func handleError(w http.ResponseWriter, _ *http.Request, err error, code int) {
	http.Error(w, err.Error(), code)
}
