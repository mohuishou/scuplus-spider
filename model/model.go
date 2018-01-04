package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/mohuishou/scuplus-spider/log"
)

var db *gorm.DB

// Model 基本模型的定义
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt int64
}

// init 数据库初始化
func initDB() {
	var err error
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", "root", "123456", "localhost", "32770", "scuplus-spider"))
	if err != nil {
		log.Fatal("数据库连接错误：", err)
	}
	db.AutoMigrate(&Detail{Model: Model{}}, &Tag{Model: Model{}}, &DetailTag{Model: Model{}})
}

func DB() *gorm.DB {
	if db == nil {
		initDB()
	}
	return db
}

// Close 关闭数据库连接
func Close() error {
	return DB().Close()
}
