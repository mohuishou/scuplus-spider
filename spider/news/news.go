package news

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://news.scu.edu.cn"
const category = "四川大学新闻网"

var urls = map[string]string{
	"川大在线": "cdzx",
}

func Spider(maxTryNum int, key string) {
	// 入口链接
	url := fmt.Sprintf("http://news.scu.edu.cn/%s.htm", urls[key])
	tryCount := 0

	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML(".winstyle195338 tr", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("td:nth-child(3)"))
		t, err := time.Parse("2006-01-02", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}
		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		go e.Request.Visit(domain + "/" + strings.Trim(e.ChildAttr("a", "href"), ".."))
	})

	// 列表页： 获取下一页
	c.OnHTML(".Next", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		if strings.Contains(e.Text, "下页") {
			c.Visit(domain + "/cdzx/" + strings.Trim(e.Attr("href"), "cdzx/"))
		}
	})

	// 获取内容页信息
	c.OnHTML("form[name=\"_newscontent_fromname\"]", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "info") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}年\d{1,2}月\d{1,2}日\s\d{2}:\d{2}`)
		createdStr := r.FindString(e.ChildText("div > div:nth-child(2)"))
		createdAt := spider.StrToTime("2006年01月02日 15:04", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("div > div:nth-child(3)")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("div > h3")

		// 获取标签
		tagIDs := spider.GetTagIDs(title, []string{category, key})

		detail := &model.Detail{
			Title:    title,
			Content:  content,
			Category: category,
			URL:      e.Request.URL.String(),
			Model:    model.Model{CreatedAt: createdAt},
		}

		detail.Create(tagIDs)
	})

	c.Visit(url)
	c.Wait()
}

func Run() {
	for k := range urls {
		Spider(config.GetConfig("").MaxTryNum, k)
	}
}
