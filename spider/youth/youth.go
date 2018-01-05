package youth

import (
	"fmt"
	"regexp"
	"time"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/config"
)

const domain = "http://youth.scu.edu.cn"

var urls = map[string]string{
	"公告":   "notice",
	"团情快讯": "news-group",
	"院系风采": "courtyard-style",
}

func Spider(conf config.Spider) {
	if _, ok := urls[conf.Key]; !ok {
		log.Fatal("[E]: 不存在这个key")
	}

	// 入口链接
	url := fmt.Sprintf("http://youth.scu.edu.cn/index.php/main/web/%s/p/1", urls[conf.Key])

	tryCount := 0

	c := colly.NewCollector()

	c.DetectCharset = true

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// 获取列表页面的所有列表
	c.OnHTML("ul.list-art li", func(e *colly.HTMLElement) {

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
	c.OnHTML(".pagination a.next", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > conf.MaxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		if e.Attr("href") != "" {
			go e.Request.Visit(domain + e.Attr("href"))
		}
	})

	// 获取内容页信息
	c.OnHTML(".content-art", func(e *colly.HTMLElement) {
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}\s\d{1,2}:\d{1,2}:\d{1,2}`)
		createdStr := r.FindString(e.ChildText(".attr"))
		createdAt := spider.StrToTime("2006-01-02 15:04", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find(".content-text")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("h1")
		// 获取标签
		tagIDs := spider.GetTagIDs(title, []string{conf.Key})

		detail := &model.Detail{
			Title:    title,
			Content:  content,
			Category: "青春川大",
			URL:      e.Request.URL.String(),
			Model:    model.Model{CreatedAt: createdAt},
		}

		detail.Create(tagIDs)
	})

	c.Visit(url)

	c.Wait()
}
