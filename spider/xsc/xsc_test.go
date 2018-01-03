package xsc

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
					IsNew:     false,
					Key:       "新闻",
					Second:    3600 * 24 * 30,
					MaxTryNum: 10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Spider(tt.args.conf)
		})
	}
}
