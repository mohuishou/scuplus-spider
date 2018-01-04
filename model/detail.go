package model

import (
	"github.com/mohuishou/scuplus-spider/log"
)

// Detail 文章
type Detail struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
	Title     string
	Content   string
	URL       string
	Category  string
}

// Create 新建一条文章记录
func (d *Detail) Create(tags []string) {
	log.Info(d.Title, d.CreatedAt, d.URL)
}
