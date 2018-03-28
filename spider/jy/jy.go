package jy

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

const domain = "http://jy.scu.edu.cn"
const category = "就业网"

var urls = map[string]string{
	"宣讲会":  "/eweb/jygl/zpfw.so?modcode=jygl_xjhxxck&subsyscode=zpfw&type=searchXjhxx&xjhType=all",
	"招聘信息": "/eweb/jygl/zpfw.so?modcode=jygl_zpxxck&subsyscode=zpfw&type=searchZpxx&xxlb=5100",
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

	cookies, err := spider.GetCookies(domain+urls[key], ".main")
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
	c.OnHTML(".z_newsl > ul > li", func(e *colly.HTMLElement) {

		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		createdStr := e.ChildText("div:nth-child(3)")
		if key == "招聘信息" {
			createdStr = e.ChildText("div:nth-child(2)")
		}
		t, err := time.Parse("2006-01-02", createdStr)
		if err != nil {
			log.Info("时间转换失败：", err.Error())
			return
		}
		if t.Unix() < detail.CreatedAt.Unix() {
			tryCount++
			return
		}

		str := e.ChildAttr("div:nth-child(1) > a", "onclick")
		r, _ := regexp.Compile(`'.+?'`)
		id := strings.Trim(r.FindString(str), "'")

		if key == "宣讲会" {
			go e.Request.Visit(fmt.Sprintf("%s/eweb/jygl/zpfw.so?modcode=jygl_xjhxxck&subsyscode=zpfw&type=viewXjhxx&id=%s&t=%s&title=%s", domain, id, createdStr, e.ChildText("div:nth-child(1) > a")))
		} else {
			go e.Request.Visit(fmt.Sprintf("%s/eweb/jygl/zpfw.so?modcode=jygl_zpxxck&subsyscode=zpfw&type=viewZpxx&id=%s&t=%s&title=%s", domain, id, createdStr, e.ChildText("div:nth-child(1) > a")))
		}
	})

	// 列表页： 获取下一页
	c.OnHTML("#pageForm > a:nth-child(4)", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		// 获取当前页码
		if e.Text == "下一页" {
			e.Request.Visit(domain + e.Attr("href"))
		}
	})

	// 获取内容页信息
	c.OnHTML("body > div.content > div > div.z_content", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if !strings.Contains(e.Request.URL.Query().Get("type"), "view") {
			return
		}

		// 获取发布时间
		createdStr := e.Request.URL.Query().Get("t")
		createdAt := spider.StrToTime("2006-01-02", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM
		spider.LinkHandle(contentDom, domain)

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.Request.URL.Query().Get("title")

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
