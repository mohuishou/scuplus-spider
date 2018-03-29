package eie

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

const domain = "http://eie.scu.edu.cn"
const category = "电子信息学院"

var urls = map[string]string{
	"新闻":    "/xyzx/xyxw.htm",
	"公告":    "/xyzx/xytz.htm",
	"本科教育":  "/rcpy/bkjy.htm",
	"研究生教育": "/rcpy/yjsjy.htm",
}

func Spider(maxTryNum int, key string) {
	tryCount := 0

	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("body > div.list.cleafix > div.dp.fr > ul > li", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("span"))
		t, err := time.Parse("2006-01-02", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}
		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		go e.Request.Visit(domain + strings.Trim(e.ChildAttr("a", "href"), ".."))
	})

	// 列表页： 获取下一页
	c.OnHTML("span.p_next > a", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		href := strings.Split(e.Attr("href"), "/")
		if len(href) == 2 {
			c.Visit(domain + strings.Trim(urls[key], ".htm") + "/" + href[1])
		}
	})

	// 获取内容页信息
	c.OnHTML("div.content_box .content", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "info") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("author"))
		createdAt := spider.StrToTime("2006-01-02", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("#vsb_content_2")
		spider.LinkHandle(contentDom, domain+"/")

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("h1")

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

	c.Visit(domain + urls[key])
	c.Wait()
}

func Run() {
	for k := range urls {
		Spider(config.GetConfig("").MaxTryNum, k)
	}
}
