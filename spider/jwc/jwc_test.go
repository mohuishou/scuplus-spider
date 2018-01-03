package jwc

import (
	"testing"

	"github.com/mohuishou/scuplus-spider/config"
)

func Test_spider(t *testing.T) {
	type args struct {
		conf config.Spider
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				conf: config.Spider{
					IsNew:     true,
					Key:       "公告",
					Second:    3600 * 24 * 10,
					MaxTryNum: 10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spider(tt.args.conf)
		})
	}
}