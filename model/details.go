package model

import (
	"github.com/mohuishou/scuplus-spider/log"
)

// Detail 文章
type Detail struct {
	Title    string `gorm:"unique_index"`
	Content  string `gorm:"type:text;"`
	URL      string
	Category string
	Tags     []Tag `gorm:"many2many:detail_tags;"`
	Model
}

// Create 新建一条文章记录
func (d *Detail) Create() []error {
	log.Info(d.Title)
	if errs := DB().Create(d).GetErrors(); errs != nil {
		log.Error(errs)
		return errs
	}
	return nil
}
