package pagination

// Meta holds offset pagination state for list pages.
type Meta struct {
	Page     int
	PerPage  int
	Total    int
	HasPrev  bool
	HasNext  bool
	PrevPage int
	NextPage int
}

const defaultPerPage = 25

// New builds pagination metadata for a list page.
func New(page, perPage, total int) Meta {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	hasPrev := page > 1
	hasNext := page*perPage < total
	return Meta{
		Page:     page,
		PerPage:  perPage,
		Total:    total,
		HasPrev:  hasPrev,
		HasNext:  hasNext,
		PrevPage: page - 1,
		NextPage: page + 1,
	}
}

// Offset returns the SQL OFFSET for page and perPage.
func Offset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	return (page - 1) * perPage
}
