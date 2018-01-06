package model

import (
	"fmt"

	"github.com/mohuishou/scuplus-spider/log"
)

// Detail 文章
type Detail struct {
	Model
	Author   string
	Title    string `gorm:"index"`
	Content  string `gorm:"type:text;"`
	URL      string
	Category string
	Tags     []Tag `gorm:"many2many:detail_tags"`
}

// Create 新建一条文章记录
func (d *Detail) Create(tagIDs []uint) error {
	tx := DB().Begin()
	if err := tx.Create(d).Error; err != nil {
		log.Error(err)
		tx.Rollback()
		return err
	}

	for _, id := range tagIDs {
		if err := tx.Create(&DetailTag{TagID: id, DetailID: d.ID}).Error; err != nil {
			log.Error(err)
			tx.Rollback()
			return err
		}
	}

	log.Info(fmt.Sprintf("%s %s 以保存", d.Title, d.CreatedAt.Format("2006-01-02")))

	tx.Commit()

	return nil
}
