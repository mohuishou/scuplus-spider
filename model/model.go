package model

import (
	"github.com/mohuishou/scuplus-spider/log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

// 基本模型的定义
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt int64
}

// Init 数据库初始化
func Init() {
	DB, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	if err ！=nil{
		log.Fatal(err)
	}
}

// Close 关闭数据库连接
func Close() error{
	return DB.Close()
}
