package scuinfo

import "testing"
import "github.com/mohuishou/scuplus-spider/config"

func Test_spider(t *testing.T) {
	spider(config.Spider{
		IsNew:     true,
		MaxTryNum: 10,
		Key:       "最近",
		Second:    60 * 10,
	})
}
