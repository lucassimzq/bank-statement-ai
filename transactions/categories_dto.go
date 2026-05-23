package transactions

type ListCategoriesResponse struct {
	Categories []*Category `json:"categories"`
}
