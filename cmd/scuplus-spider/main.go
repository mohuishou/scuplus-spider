package main

import (
	"sync"

	"github.com/robfig/cron"

	"github.com/mohuishou/scuplus-spider/config"
	"github.com/mohuishou/scuplus-spider/spider/jwc"
	"github.com/mohuishou/scuplus-spider/spider/news"
	"github.com/mohuishou/scuplus-spider/spider/sau"
	"github.com/mohuishou/scuplus-spider/spider/scuinfo"
	"github.com/mohuishou/scuplus-spider/spider/xsc"
	"github.com/mohuishou/scuplus-spider/spider/youth"
)

var waitgroup sync.WaitGroup

func main() {
	c := cron.New()
	c.AddFunc(config.GetConfig("").Spec, func() {
		run(config.GetConfig("").Spider)
	})
	c.Start()
	select {}
}

func run(conf config.Spider) {
	for k := range jwc.GetURLs() {
		conf.Key = k
		waitgroup.Add(1)
		go func(conf config.Spider) {
			jwc.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	for k := range news.GetURLs() {
		waitgroup.Add(1)
		conf.Key = k
		go func(conf config.Spider) {
			news.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	for k := range sau.GetURLs() {
		waitgroup.Add(1)
		conf.Key = k
		go func(conf config.Spider) {
			sau.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	for k := range scuinfo.GetURLs() {
		waitgroup.Add(1)
		conf.Key = k
		go func(conf config.Spider) {
			scuinfo.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	for k := range xsc.GetURLs() {
		waitgroup.Add(1)
		conf.Key = k
		go func(conf config.Spider) {
			xsc.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	for k := range youth.GetURLs() {
		conf.Key = k
		waitgroup.Add(1)
		go func(conf config.Spider) {
			youth.Spider(conf)
			waitgroup.Done()
		}(conf)
	}

	waitgroup.Wait()
}
