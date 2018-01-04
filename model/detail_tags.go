package model

// DetailTag .
type DetailTag struct {
	Model
	DetailID uint `gorm:"primary_key:no"`
	TagID    uint `gorm:"primary_key:no"`
}
