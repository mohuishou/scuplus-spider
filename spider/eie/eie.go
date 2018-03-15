package eie

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

const domain = "http://eie.scu.edu.cn"
const category = "电子信息学院"

var urls = map[string]string{
	"新闻":    "1",
	"公告":    "2",
	"本科教育":  "9",
	"研究生教育": "10",
}

func Spider(maxTryNum int, key string) {
	// 入口链接
	url := fmt.Sprintf("%s/NewsList.aspx?tid=%s", domain, urls[key])
	tryCount := 0

	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)
	// 获取列表页面的所有列表
	c.OnHTML(".onetitle", func(e *colly.HTMLElement) {
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
		go e.Request.Visit(domain + "/" + strings.Trim(e.ChildAttr("a", "href"), ".."))
	})

	// 列表页： 获取下一页
	c.OnHTML(".yahoo a:last-child", func(e *colly.HTMLElement) {
		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}
		c.Visit(domain + "/NewsList.aspx" + e.Attr("href"))
	})

	// 获取内容页信息
	c.OnHTML(".contentlist", func(e *colly.HTMLElement) {
		//判断是否是内容页
		if strings.Contains(e.Request.URL.Path, "tid") {
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}-\d{1,2}-\d{1,2}`)
		createdStr := r.FindString(e.ChildText("newsinfo"))
		createdAt := spider.StrToTime("2006-01-02日", createdStr)

		// content 替换链接 a,img
		contentDom := e.DOM.Find(".newsdesc")
		spider.LinkHandle(contentDom, domain+"/")

		// 获取正文
		content, err := contentDom.Html()
		if err != nil {
			log.Error("获取内容页失败：", err.Error())
			return
		}

		// 获取标题
		title := e.ChildText(".newstitle")

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

// GetURLs 获取所有的url
func GetURLs() map[string]string {
	return urls
}
