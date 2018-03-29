package jwc

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/PuerkitoBio/goquery"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://jwc.scu.edu.cn/jwc/"
const category = "教务处"

var urls = map[string]string{
	"公告": "moreNotice",
	"新闻": "moreNews",
}

func Spider(maxTryNum int, key string) {
	tryCount := 0
	c := spider.NewCollector()
	cookies, err := spider.GetCookies(domain+urls[key], "table")
	if err != nil {
		log.Warn("cookie获取错误", err)
		return
	}
	err = c.SetCookies(domain, cookies)
	if err != nil {
		log.Warn("cookie设置错误", err)
		return
	}
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML("#news_list > table > tbody > tr", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
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

		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}

		// 发现内容页链接
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML("#pagenext", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取当前页码
		pageNow, err := strconv.Atoi(e.Attr("value"))
		if err != nil {
			log.Error("当前页码获取错误：", err.Error())
		}

		// 发现下一列表页
		go e.Request.Visit(fmt.Sprintf("%s%s.action?url=%s.action&type=2&keyWord=&pager.pageNow=%d", domain, urls[key], urls[key], pageNow+1))
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
		createdAt := spider.StrToTime("2006.01.02", createdStr)

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

	c.Visit(fmt.Sprintf("http://jwc.scu.edu.cn/jwc/%s.action", urls[key]))
	c.Wait()
}

func Run() {
	for k := range urls {
		Spider(config.GetConfig("").MaxTryNum, k)
	}
}
