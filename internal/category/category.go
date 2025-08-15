package category

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
	// Category string `json:"category"`
	Stock int `json:"stock"`
}
