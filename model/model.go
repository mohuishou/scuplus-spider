package model

import (
	"fmt"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // mysql驱动
	"github.com/mohuishou/scuplus-spider/log"
)

var db *gorm.DB

// Model 基本模型的定义
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// init 数据库初始化
func initDB() {
	// 获取配置
	conf := config.GetConfig().Mysql

	// 初始化连接
	var err error
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", conf.User, conf.Password, conf.Host, conf.Port, conf.DB))
	if err != nil {
		log.Fatal("数据库连接错误：", err)
	}
	db.AutoMigrate(&DetailTag{}, &Detail{}, &Tag{})
}

// DB 返回db，如果不存在则初始化
func DB() *gorm.DB {
	if db == nil {
		initDB()
	}
	db.LogMode(true)
	return db
}

// Close 关闭数据库连接
func Close() error {
	return DB().Close()
}
