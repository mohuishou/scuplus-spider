package youth

import (
	"fmt"
	"regexp"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"
)

const domain = "http://youth.scu.edu.cn"
const category = "青春川大"

var urls = map[string]string{
	"公告":   "notice",
	"团情快讯": "news-group",
	"院系风采": "courtyard-style",
}

func Spider(maxTryNum int, key string) {
	tryCount := 0

	c := spider.NewCollector()
	// 获取最后一条数据的时间
	detail := model.GetLastDetail(category, key)

	// 获取列表页面的所有列表
	c.OnHTML("ul.list-art li", func(e *colly.HTMLElement) {

		// 判断是否为最新的页面，如果不是则丢弃
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		t, err := time.Parse("2006-01-02", e.ChildText("span"))
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
	c.OnHTML(".pagination a.next", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
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
		createdAt := spider.StrToTime("2006-01-02 15:04:05", createdStr)

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
		tagIDs := spider.GetTagIDs(title, []string{key, category})

		detail := &model.Detail{
			Title:    title,
			Content:  content,
			Category: category,
			URL:      e.Request.URL.String(),
			Model:    model.Model{CreatedAt: createdAt},
		}

		detail.Create(tagIDs)
	})

	c.Visit(fmt.Sprintf("http://youth.scu.edu.cn/index.php/main/web/%s/p/1", urls[key]))
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
