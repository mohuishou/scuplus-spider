package scuinfo

import "testing"
import "github.com/mohuishou/scuplus-spider/config"

func Test_spider(t *testing.T) {
	Spider(config.Spider{
		IsNew:     false,
		MaxTryNum: 10,
		Key:       "热门",
		Second:    60 * 10,
	})
}
