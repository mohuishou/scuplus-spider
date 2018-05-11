package sau

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

const domain = "http://sau.scu.edu.cn"
const category = "社团联"

// 最大页码，由于全量数据一般只执行一次，所以直接写死
const maxPage = 226

var urls = map[string]string{
	"公告": "/chronicle/notice",
	"新闻": "/club/clubnews",
}

func Spider(maxTryNum int, key string) {

	tryCount := 0
	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("body > table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(3) > td > table  tr", func(e *colly.HTMLElement) {

		if e.ChildText("td:nth-child(2)") == "" {
			return
		}

		// 判断是否为最新的页面，如果不是则丢弃

		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		t, err := time.Parse("2006-01-02", e.ChildText("td:nth-child(2)"))
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
	c.OnHTML("#badoopager", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
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
			go e.Request.Visit(fmt.Sprintf("%s?pageid=%s", domain+urls[key], s[0][1]))
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
		createdAt := spider.StrToTime("2006-01-02", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(6) > td > div")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(2) > td > span")
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
