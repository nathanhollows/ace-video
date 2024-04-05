package models

type JSONOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
	Search string
	Tags   string
	Type   string
	ID     string
}
