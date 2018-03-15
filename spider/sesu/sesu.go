package sesu

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/mohuishou/scuplus-spider/spider"

	"github.com/PuerkitoBio/goquery"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"

	"github.com/chromedp/chromedp/client"

	"github.com/chromedp/cdproto/network"

	"github.com/chromedp/cdproto/cdp"

	"github.com/chromedp/chromedp"
)

const domain = "http://sesu.scu.edu.cn"
const category = "经济学院"

var urls = map[string]string{
	"新闻":    "/news",
	"公告":    "/gonggao",
	"本科教育":  "/personnelTraining/bachelor",
	"研究生教育": "/personnelTraining/graduate",
}

func Spider(maxTryNum int, key string) {
	tryCount := 0
	c := spider.NewCollector()
	cookies, err := GetCookies()
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
	c.OnHTML(".news_list.r ul li", func(e *colly.HTMLElement) {
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}\-\d{1,2}\-\d{1,2}`)
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

		// 发现内容页链接
		go e.Request.Visit(domain + e.ChildAttr("a", "href"))
	})

	// 列表页： 获取下一页
	c.OnHTML(".pagelist li a", func(e *colly.HTMLElement) {

		// 如果仅需获取最新内容，判断是否已经达到最大尝试次数
		if tryCount > maxTryNum {
			log.Info("已达到最大尝试次数")
			return
		}

		// 获取当前页码
		if strings.Contains(e.Text, "下一页") {
			// 发现下一列表页
			go e.Request.Visit(fmt.Sprintf("%s%s/%s", domain, urls[key], e.Attr("href")))
		}

	})

	// 获取内容页信息
	c.OnHTML(".news_list.r", func(e *colly.HTMLElement) {
		if e.DOM.Find(".mt20").Text() == "" {
			return
		}
		// 获取发布时间
		r, _ := regexp.Compile(`\d{4}\-\d{1,2}\-\d{1,2}\s\d{1,2}:\d{1,2}`)
		createdStr := r.FindString(e.ChildText(".info"))
		createdAt := spider.StrToTime("2006-01-02 15:04", createdStr)

		// 获取正文
		content := e.ChildAttr(".mt20 > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1)", "value")
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
		title := e.ChildText(".view_title")

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

	c.Visit(fmt.Sprintf(domain + urls[key]))
	c.Wait()
}

func Run() {
	for k := range urls {
		Spider(config.GetConfig("").MaxTryNum, k)
	}
}

// GetCookies 获取cookie字符串
func GetCookies() ([]*http.Cookie, error) {
	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithTargets(client.New().WatchPageTargets(ctxt)))
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	cookies := []*http.Cookie{}

	// run task list
	err = c.Run(ctxt, chromedp.Tasks{
		chromedp.Navigate(domain),
		chromedp.WaitVisible(`.content`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
			allCookies, err := network.GetAllCookies().Do(ctx, h)
			for _, v := range allCookies {
				cookies = append(cookies, &http.Cookie{
					Name:   v.Name,
					Value:  v.Value,
					Domain: v.Domain,
					Path:   v.Path,
				})
			}
			if err != nil {
				return err
			}
			return nil
		}),
	})
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	return cookies, nil
}

// GetURLs 获取所有的url
func GetURLs() map[string]string {
	return urls
}
