package model

type Tag struct {
	Name string `gorm:"unique_index"`
	Model
}
