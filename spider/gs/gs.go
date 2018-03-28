package gs

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

const domain = "http://gs.scu.edu.cn"
const category = "研究生院"

var urls = map[string]string{
	"公告": "/NewNotice.aspx",
	"新闻": "/News.aspx",
	"招生": "/zhaosheng.aspx",
}

var params = map[string]string{}

func Spider(maxTryNum int, key string) {
	tryCount := 0

	c := spider.NewCollector()

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       2 * time.Second, // 随机延时
	})

	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	c.OnHTML(`html`, func(e *colly.HTMLElement) {
		if e.ChildText("body > div > div:nth-child(1)") == "网站访问认证，点击链接后将跳转到访问页面" {
			str := e.ChildText("head script")
			reg, _ := regexp.Compile(`".+?"`)
			res := reg.FindString(str)
			log.Info("安全狗认证", domain+strings.Trim(res, `"`))
			c.Visit(domain + strings.Trim(res, `"`))
			c.Visit(domain + urls[key])
		}
	})
	c.OnHTML("#form1 input", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if strings.Contains(e.Request.URL.Path, "newDetail") {
			return
		}
		params[e.Attr("name")] = e.Attr("value")
	})
	// 获取列表页面的所有列表
	c.OnHTML("#form1 > div.warp.content > div.child_r > div.child_r_content > div > div.ListLeft > div", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.Text)
		t, err := time.Parse("2006-1-2", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}
		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}
		go e.Request.Visit(domain + "/" + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML("#ContentPlaceHolder1_AspNetPager1 > a", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取当前页码
		if e.Text == "下一页" {
			href := e.Attr("href")
			reg, _ := regexp.Compile(`'.+?'`)
			args := reg.FindAllString(href, -1)
			if len(args) == 2 {
				params["__EVENTTARGET"] = strings.Trim(args[0], "'")
				params["__EVENTARGUMENT"] = strings.Trim(args[1], "'")
				e.Request.Post(domain+urls[key], params)
			}

		}
	})

	// 获取内容页信息
	c.OnHTML("#form1 > div.warp.content > div.child_r > div.child_r_content > div:nth-child(1)", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Path, "newDetail") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}/\d{1,2}/\d{1,2}\s\d{1,2}:\d{1,2}:\d{1,2}`)
		createdStr := r.FindString(e.ChildText("#ContentPlaceHolder1_shijian"))
		createdAt := spider.StrToTime("2006/1/2 15:04:05", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find("#ContentPlaceHolder1_content")
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText("#ContentPlaceHolder1_title")

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
