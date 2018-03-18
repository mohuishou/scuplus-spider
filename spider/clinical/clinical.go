package clinical

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

const domain = "http://www.cd120.com"
const category = "华西临床医学院"

var urls = map[string]string{
	"新闻":   "/htmlnewszhongyaoxinwen/index.jhtml",
	"院内动态": "/htmlnewsdongtaixinwen/index.jhtml",
	"公告":   "/htmlnewsgonggaolan/index.jhtml",
	"院内通知": "/htmlnewsyuannatongzhi/index.jhtml",
	"会议通知": "/htmlnewshuiyizhuanlan/index.jhtml",
}

func Spider(maxTryNum int, key string) {
	// 入口链接
	tryCount := 0
	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML(".infoContent > ul:nth-child(1) > li", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取发布时间
		createdStr := e.ChildText("p:nth-child(2)")
		t, err := time.Parse("2006-01-02", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}
		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML(".pages > div:nth-child(1) > a:nth-child(3)", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		c.Visit(domain + strings.Trim(urls[key], "index.jhtml") + e.Attr("href"))
	})

	// 获取内容页信息
	c.OnHTML(".w756", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if e.ChildText("h5") == "" {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("div.author:nth-child(2)"))
		createdAt := spider.StrToTime("2006-01-02", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find(".newsContent")

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("h5:nth-child(1)")

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
