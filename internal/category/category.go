package category

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Stock    int    `json:"stock"`
}

type Repository interface {
	GetCategoryByName(n string) (*Category, error)
	GetCategoryByID(id int) (*Category, error)
	AddNewCategory(n string) (int64, error)
	UpdateCategory(n string, id int) (int64, error)
	DeleteCategory(id int) (int64, error)
}
