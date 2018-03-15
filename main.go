package main

import (
	"github.com/mohuishou/scuplus-spider/spider/news"
	"github.com/mohuishou/scuplus-spider/spider/youth"
	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	// 每两小时抓取一次
	c.AddFunc("@every 2h", func() {
		// 四川大学新闻网
		news.Run()
		// 青春川大
		youth.Run()
	})
	c.Start()
	select {}
}
