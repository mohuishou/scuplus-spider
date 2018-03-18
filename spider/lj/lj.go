// Package lj 文学与新闻学院
package lj

import (
	"regexp"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://lj.scu.edu.cn"
const category = "文学与新闻学院"

var urls = map[string]string{
	"公告": "/cdwxxy1.2/index.php?app=default&act=article&id=0580F3BF8A8794F53FA932DF5C79AFB1",
	"新闻": "/cdwxxy1.2/index.php?app=default&act=article&id=C87B31A5B538412CFCAA01570A8645D8",
}

func Spider(maxTryNum int, key string) {

	tryCount := 0
	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("div.artiFz:nth-child(4) > ul > li", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		t, err := time.Parse("2006-01-02", e.ChildText("span.am-fr"))
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}

		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		// 发现内容页链接
		go e.Request.Visit(e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML(".am-pagination-next a", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 发现下一列表页
		go e.Request.Visit(e.Attr("href"))
	})

	// 获取内容页信息
	c.OnHTML("div.bkWt", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if e.Request.URL.Query().Get("act") != "article_view" {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}\s\d{1,2}:\d{1,2}:\d{1,2}`)
		createdStr := r.FindString(e.ChildText("p.ft5"))
		createdAt := spider.StrToTime("2006-01-02 15:04:05", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("div.ft5")
		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText(".borBottom > h2:nth-child(1)")
		// 获取标签
		tagsIDs := spider.GetTagIDs(title, []string{category, key})

		detail := &model.Detail{
			Title:    title,
			Content:  content,
			Category: category,
			URL:      e.Request.URL.String(),
			Model:    model.Model{CreatedAt: createdAt},
		}

		detail.Create(tagsIDs)
	})
	c.Visit(domain + urls[key])
	c.Wait()
}

func Run() {
	for k := range urls {
		Spider(config.GetConfig("").MaxTryNum, k)
	}
}
