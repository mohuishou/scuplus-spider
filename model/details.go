package model

import (
	"errors"
	"fmt"

	"github.com/mohuishou/scuplus-spider/log"
)

// Detail 文章
type Detail struct {
	Model
	Author   string
	Title    string `gorm:"index"`
	Content  string `gorm:"type:longtext;"`
	URL      string
	Category string
	Tags     []Tag `gorm:"many2many:detail_tags"`
}

// Create 新建一条文章记录
func (d *Detail) Create(tagIDs []uint) error {
	if d.Title == "" || d.Content == "" {
		log.Error("标题内容不能为空")
		return errors.New("标题内容不能为空")
	}
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

	log.Info(fmt.Sprintf("%s %s 已保存", d.Title, d.CreatedAt.Format("2006-01-02")))

	tx.Commit()

	return nil
}

// GetLastDetail 获取某些标签下的最后一篇文章
func GetLastDetail(category, name string) *Detail {
	tag, ok := Tags[name]

	if !ok {
		t := Tag{Name: name}
		DB().Create(&t)
		tag = t
	}
	detail := Detail{}
	DB().Debug().Model(&tag).Where("category = ?", category).Order("details.created_at desc").Limit(1).Related(&detail, "Details")
	return &detail
}
