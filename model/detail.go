package model

type Detail struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
	Title     string
	Content   string
	URL       string
	Category  string
}

func (d *Detail) Create() {

}
