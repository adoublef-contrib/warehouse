package category

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	categoryService *Service
}

func NewCategoryHandler(s *Service) *Handler {
	return &Handler{categoryService: s}
}

type ErrorMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type SuccessMessage struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    []*Category `json:"data"`
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorMessage{
		Message: message,
		Code:    statusCode,
	}

	// Create a function to handle this altogether
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		log.Printf("failed to encode json: %v", err)
	}
}
func (h *Handler) GetCategoryByName(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	category := params.Get("category")

	if category == "" {
		sendJSONError(w, "Category cannot be blank", http.StatusBadRequest)
	}

	cat, err := h.categoryService.GetCategoryByName(category)

	if err != nil {
		log.Printf("GetCategoryByName error: %v", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := SuccessMessage{
		Message: "Successfully fetched category",
		Code:    http.StatusOK,
		Data:    []*Category{cat},
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	id, err := strconv.Atoi(params.Get("id"))

	if err != nil {
		log.Printf("Failed to convert id to int: %v", err)
	}

	cat, err := h.categoryService.GetCategoryByID(id)

	if err != nil {
		log.Printf("GetCategoryByID error: %v", err)
		sendJSONError(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	data := SuccessMessage{
		Message: "Successfully retrieved data by ID",
		Code:    http.StatusOK,
		Data:    []*Category{cat},
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode json: %v", err)
	}

}

func (h *Handler) AddNewCategory(w http.ResponseWriter, r *http.Request) {
	var newCat string
	err := json.NewDecoder(r.Body).Decode(&newCat)

	if err != nil {
		log.Printf("Error decoding body: %v", err)

	}

	id, err := h.categoryService.AddNewCategory(newCat)

	if err != nil {
		log.Printf("AddNewCategoryError: %v", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]any{"message": "Successfully created new category", "code": http.StatusCreated, "name": newCat, "id": id}); err != nil {
		log.Printf("failed to encode json: %v", err)
	}
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var cat *Category

	err := json.NewDecoder(r.Body).Decode(&cat)

	if err != nil {
		log.Printf("Error decoding body: %v", err)
		return
	}

	id, err := h.categoryService.UpdateCategory(cat.Name, cat.ID)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "Category not found", http.StatusBadRequest)
			log.Printf("No rows were found with the specified category")
			return
		} else {
			log.Printf("UpdateCategory error: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]any{"message": "Successfully updated category", "code": http.StatusOK, "newCat": cat.Name, "id": id}); err != nil {
		fmt.Printf("failed to encode json: %v", err)
	}
}

// func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
// 	var cat Category
//
// 	err := json.NewDecoder(r.Body).Decode(&cat)
//
// 	if err != nil {
// 		log.Printf("Error decoding body: %v", err)
// 		return
// 	}
//
// 	id, err := h.categoryService.DeleteCategory(cat.CategoryID)
//
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			log.Printf("Now rows were found with this specified id")
// 			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
// 			return
// 		} else {
// 			log.Printf("DeleteCategory error: %v", err)
// 			sendJSONError(w, "Internal server error", http.StatusInternalServerError)
// 			return
// 		}
// 	}
//
// 	w.Header().Set("Content-type", "application/json")
//
// 	if err := json.NewEncoder(w).Encode(map[string]any{"message": "Successfully deleted category", "code": http.StatusOK, "category": id}); err != nil {
// 		log.Printf("Failed to encode JSON: %v", err)
// 	}
// }
