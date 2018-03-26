package gs

import (
	"testing"

	"github.com/mohuishou/scuplus-spider/config"
)

func TestRun(t *testing.T) {
	Spider(config.GetConfig("").MaxTryNum, "新闻")
}
