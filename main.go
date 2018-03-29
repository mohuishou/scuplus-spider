package main

import (
	"github.com/mohuishou/scuplus-spider/spider/clinical"
	"github.com/mohuishou/scuplus-spider/spider/cs"
	"github.com/mohuishou/scuplus-spider/spider/eie"
	"github.com/mohuishou/scuplus-spider/spider/gs"
	"github.com/mohuishou/scuplus-spider/spider/jwc"
	"github.com/mohuishou/scuplus-spider/spider/jy"
	"github.com/mohuishou/scuplus-spider/spider/lj"
	"github.com/mohuishou/scuplus-spider/spider/news"
	"github.com/mohuishou/scuplus-spider/spider/sau"
	"github.com/mohuishou/scuplus-spider/spider/seei"
	"github.com/mohuishou/scuplus-spider/spider/sesu"
	"github.com/mohuishou/scuplus-spider/spider/xsc"
	"github.com/mohuishou/scuplus-spider/spider/youth"
	"github.com/robfig/cron"
)

func main() {
	c := cron.New()
	// 从早上七点开始每两小时抓取一次
	c.AddFunc("0 0 7/2 * * ?", func() {
		// 青春川大
		youth.Run()
		// 社团联
		sau.Run()
		// 教务处
		jwc.Run()
		// 学工部
		xsc.Run()
		// 四川大学新闻网
		news.Run()
	})
	// 从早上七点半开始每两小时抓取一次
	c.AddFunc("0 30 7/2 * * ? ", func() {
		// 研究生院
		gs.Run()
		// 就业网
		jy.Run()
		// 电子信息学院
		eie.Run()
		// 电气信息学院
		seei.Run()

	})
	// 从早上八点开始每两小时抓取一次
	c.AddFunc("0 0 8/2 * * ? ", func() {
		// 经济学院
		sesu.Run()
		// 计算机学院
		cs.Run()
		//文新
		lj.Run()
		// 华西临床医学院
		clinical.Run()
	})
	c.Start()
	select {}
}
