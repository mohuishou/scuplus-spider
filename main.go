package main

import (
	"github.com/mohuishou/scuplus-spider/spider/news"
	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	// 每两小时抓取一次
	c.AddFunc("@every 2h", func() {
		// 四川大学新闻网
		news.Run()
		// 教务处
	})
	c.Start()
	select {}
}
