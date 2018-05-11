package jwc

import (
	"regexp"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://jwc.scu.edu.cn/"
const category = "教务处"

var urls = map[string]string{
	"公告": "index.htm",
}

func Spider(maxTryNum int, key string) {
	tryCount := 0
	c := spider.NewCollector()

	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("#serach_div > div:nth-child(2) > div > div.list-llb-box > ul > li", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		createdStr := e.ChildText("a em.list-date-a")
		t, err := time.Parse("[2006.01.02]", createdStr)
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
	// c.OnHTML("#serach_div > div.list-a-content > div.list-b-right.fr > div > div > a:nth-child(7)", func(e *colly.HTMLElement) {

	// 	// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
	// 	if tryCount > maxTryNum {
	// 		log.Info("已达到最大尝试次数")
	// 		return
	// 	}

	// 	// 发现下一列表页
	// 	go e.Request.Visit(e.Attr("href"))
	// })

	// 获取内容页信息
	c.OnHTML("#serach_div > div.list-a-content", func(e *colly.HTMLElement) {
		// 获取标题
		title := strings.TrimSpace(e.ChildText(".page-title"))

		//判断是否是内容页
		if title == "" {
			log.Info("非内容页，丢弃")
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}\-\d{1,2}\-\d{1,2}`)
		createdStr := r.FindString(e.ChildText(" p > span:nth-child(1)"))
		createdAt := spider.StrToTime("2006-01-02", createdStr)

		// 获取正文
		content, _ := e.DOM.Find("div.page-content").Html()
		attachment, _ := e.DOM.Find("div.fj-main").Html()
		content = content + attachment

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
