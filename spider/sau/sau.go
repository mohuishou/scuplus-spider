package sau

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/config"
)

const domain = "http://sau.scu.edu.cn"

// 最大页码，由于全量数据一般只执行一次，所以直接写死
const maxPage = 226

var urls = map[string]string{
	"公告": "/chronicle/notice",
	"新闻": "/club/clubnews",
}

func Spider(conf config.Spider) {
	if _, ok := urls[conf.Key]; !ok {
		log.Fatal("[E]: 不存在这个key")
	}

	// 入口链接
	url := domain + urls[conf.Key]

	tryCount := 0

	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// 获取列表页面的所有列表
	c.OnHTML("body > table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(3) > td > table > tbody > tr", func(e *colly.HTMLElement) {

		if e.ChildText("td:nth-child(2)") == "" {
			return
		}

		// 判断是否为最新的页面，如果不是则丢弃
		if conf.IsNew {

			if tryCount > conf.MaxTryNum {
				log.Info("已达到最大尝试次数")
				return
			}

			// 获取发布时间
			t, err := time.Parse("2006-01-02", e.ChildText("td:nth-child(2)"))
			if err != nil {
				log.Info("时间转换失败：", err.Error())
				return
			}

			if time.Now().Unix()-t.Unix() > int64(conf.Second) {
				log.Info("数据已过期，即将被丢弃：", e.ChildText("a"))
				tryCount++
				return
			}
		}

		// 发现内容页链接
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML("#badoopager", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > conf.MaxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		pageHTML, err := e.DOM.Html()
		if err != nil {
			log.Error("页码获取错误：", err)
			return
		}

		// 发现下一列表页
		r, _ := regexp.Compile(`(\d+)">下一页</a>`)
		s := r.FindAllStringSubmatch(pageHTML, -1)
		if s != nil && len(s[0]) == 2 {
			go e.Request.Visit(fmt.Sprintf("%s?pageid=%s", url, s[0][1]))
		}
	})

	// 获取内容页信息
	c.OnHTML("body", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "newsDetail") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(2) > td > p"))
		createdAt := int64(0)
		if createdStr != "" {
			t, err := time.Parse("2006-01-02", createdStr)
			if err != nil {
				log.Error("时间转换失败：", err.Error())
			}
			createdAt = t.Unix()
		}

		// content 替换链接 a,img
		contentDom := e.DOM.Find("table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(6) > td > div")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		detail := &model.Detail{
			Title:     e.ChildText("table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(2) > td > span"),
			Content:   content,
			Category:  "社团联",
			URL:       e.Request.URL.String(),
			CreatedAt: createdAt,
		}

		detail.Create()
	})

	c.Visit(url)

	c.Wait()
}
