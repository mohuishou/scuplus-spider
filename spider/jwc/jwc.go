package jwc

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/PuerkitoBio/goquery"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/config"
)

const domain = "http://jwc.scu.edu.cn/jwc/"

// 最大页码，由于全量数据一般只执行一次，所以直接写死
const maxPage = 226

var urls = map[string]string{
	"公告": "moreNotice",
	"新闻": "moreNews",
}

func Spider(conf config.Spider) {
	if _, ok := urls[conf.Key]; !ok {
		log.Fatal("[E]: 不存在这个key")
	}

	// 入口链接
	url := fmt.Sprintf("http://jwc.scu.edu.cn/jwc/%s.action", urls[conf.Key])

	tryCount := 0

	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// 获取列表页面的所有列表
	c.OnHTML("#news_list > table > tbody > tr", func(e *colly.HTMLElement) {

		// 判断是否为最新的页面，如果不是则丢弃
		if conf.IsNew {

			if tryCount > conf.MaxTryNum {
				log.Info("已达到最大尝试次数")
				return
			}

			// 获取发布时间
			r, _ := regexp.Compile(`\d{4}\.\d{1,2}\.\d{1,2}`)
			createdStr := r.FindString(e.ChildText("td:nth-child(2)"))
			t, err := time.Parse("2006.01.02", createdStr)
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
	c.OnHTML("#pagenext", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > conf.MaxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取当前页码
		pageNow, err := strconv.Atoi(e.Attr("value"))
		if err != nil {
			log.Error("当前页码获取错误：", err.Error())
		}

		// 发现下一列表页
		if pageNow < maxPage {
			go e.Request.Visit(fmt.Sprintf("%s%s.action?url=%s.action&type=2&keyWord=&pager.pageNow=%d", domain, urls[conf.Key], urls[conf.Key], pageNow+1))
		}

	})

	// 获取内容页信息
	c.OnHTML("body", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "newsShow") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}\.\d{1,2}\.\d{1,2}`)
		createdStr := r.FindString(e.ChildText("table:nth-child(3) > tbody > tr:nth-child(4) > td"))
		createdAt := int64(0)
		if createdStr != "" {
			t, err := time.Parse("2006.01.02", createdStr)
			if err != nil {
				log.Error("时间转换失败：", err.Error())
			}
			createdAt = t.Unix()
		}

		// 获取正文
		content := e.ChildAttr("#news_content", "value")
		content = html.UnescapeString(content)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Error("正文获取错误：", err)
			return
		}
		spider.LinkHandle(doc.Selection, domain)
		content, err = doc.Html()
		if err != nil {
			log.Error("正文获取错误：", err)
			return
		}

		// 获取标题
		title := e.ChildText("table:nth-child(3) > tbody > tr:nth-child(2) > td > b")

		// 获取标签
		tags := spider.GetTag(title, []string{conf.Key})

		detail := &model.Detail{
			Title:    title,
			Content:  content,
			Category: "教务处",
			URL:      e.Request.URL.String(),
			Model:    model.Model{CreatedAt: createdAt},
			Tags:     tags,
		}

		detail.Create()
	})

	c.Visit(url)

	c.Wait()
}
