package cs

import (
	"regexp"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://cs.scu.edu.cn"
const category = "计算机学院"

var urls = map[string]string{
	"公告": "/cs/xytz/H9502index_1.htm",
	"新闻": "/cs/xyxw/H9501index_1.htm",
}

func Spider(maxTryNum int, key string) {

	tryCount := 0
	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("body > table:nth-child(11) > tbody:nth-child(1) > tr:nth-child(2) > td:nth-child(1) > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > table:nth-child(2) > tbody:nth-child(1) > tr", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("div"))
		t, err := time.Parse("2006-01-02", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}

		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		// 发现内容页链接
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML(".hangjc > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > a", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		if strings.Contains(e.Text, "下一页") {
			go e.Request.Visit(domain + e.Attr("href"))
		}
	})

	// 获取内容页信息
	c.OnHTML("body", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "webinfo") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}\s\d{1,2}:\d{1,2}`)
		createdStr := r.FindString(e.ChildText("body > table:nth-child(11) > tbody:nth-child(1) > tr:nth-child(2) > td:nth-child(1) > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1)"))
		createdAt := spider.StrToTime("2006-01-02 15:04", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("#BodyLabel")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText(".pcenter_t2 > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1)")
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

// GetURLs 获取所有的url
func GetURLs() map[string]string {
	return urls
}
