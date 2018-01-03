package xsc

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/config"
)

const domain = "http://xsc.scu.edu.cn"

var urls = map[string]string{
	"公告": "43",
	"新闻": "42",
}

func spider(conf config.Spider) {
	if _, ok := urls[conf.Key]; !ok {
		log.Fatal("[E]: 不存在这个key")
	}

	// 入口链接
	url := fmt.Sprintf("http://xsc.scu.edu.cn/P/PartialArticle?id=%s&menu=%s&rn=1", urls[conf.Key], urls[conf.Key])

	tryCount := 0

	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// 获取列表页面的所有列表
	c.OnHTML("body > ul > li", func(e *colly.HTMLElement) {

		// 判断是否为最新的页面，如果不是则丢弃
		if conf.IsNew {

			if tryCount > conf.MaxTryNum {
				log.Info("已达到最大尝试次数")
				return
			}

			t, err := time.Parse("2006-01-02", e.ChildText("span"))
			if err != nil {
				log.Info("时间转换失败：", err.Error())
				return
			}

			if time.Now().Unix()-t.Unix() > int64(conf.Second) {
				log.Info("数据已过期，即将被丢弃：", e.Text)
				tryCount++
				return
			}
		}

		// 发现内容页链接
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML(".u-list-footer a.cur", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > conf.MaxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		pageNo, err := strconv.Atoi(e.Text)
		if err != nil {
			log.Error("页码获取失败：", err.Error())
			return
		}

		if e.Attr("href") != "" {
			go e.Request.Visit(fmt.Sprintf("http://xsc.scu.edu.cn/P/PartialArticle?id=%s&menu=%s&rn=%d", urls[conf.Key], urls[conf.Key], pageNo+1))
		}
	})

	// 获取内容页信息
	c.OnHTML("body > section:nth-child(5)", func(e *colly.HTMLElement) {
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("div.v-info-info"))
		createdAt := int64(0)
		if createdStr != "" {
			t, err := time.Parse("2006-01-02", createdStr)
			if err != nil {
				log.Error("时间转换失败：", err.Error())
			}
			createdAt = t.Unix()
		}

		// 获取正文
		content, err := e.DOM.Find("div.v-info-content").Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
		}

		detail := &model.Detail{
			Title:     e.ChildText("h1"),
			Content:   content,
			Category:  "学工部",
			URL:       e.Request.URL.String(),
			CreatedAt: createdAt,
		}

		detail.Create()
	})

	c.Visit(url)

	c.Wait()
}
