package sau

import (
	"testing"

	"github.com/mohuishou/scuplus-spider/config"
)

func Test_spider(t *testing.T) {
	// Run()
	Spider(config.GetConfig("").MaxTryNum, "社团活动")
}
