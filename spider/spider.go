package spider

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mohuishou/scuplus-spider/config"

	"github.com/PuerkitoBio/goquery"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/gocolly/colly"

	"github.com/chromedp/chromedp/client"

	"github.com/chromedp/cdproto/network"

	"github.com/chromedp/cdproto/cdp"

	"github.com/chromedp/chromedp"
)

// NewCollector 新建一个Collector
func NewCollector() *colly.Collector {
	c := colly.NewCollector()
	c.DetectCharset = true
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	return c
}

// NewCollectorWithCookie 新建一个通过认证的Collector
func NewCollectorWithCookie(domain, tag string) *colly.Collector {
	c := colly.NewCollector()
	c.DetectCharset = true
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	cookies, err := GetCookies("http://"+domain, tag)
	if err != nil {
		log.Warn("cookie获取错误", err)
		return nil
	}
	err = c.SetCookies(domain, cookies)
	if err != nil {
		log.Warn("cookie设置错误", err)
		return nil
	}
	return c
}

// LinkHandle 替换img/a标签链接
func LinkHandle(contentDom *goquery.Selection, domain string) {
	contentDom.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exist := s.Attr("src"); exist && !strings.Contains(src, "http") && !strings.Contains(src, "base64") {
			s.SetAttr("src", domain+src)
		}
	})
	contentDom.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exist := s.Attr("href"); exist && !strings.Contains(href, "http") {
			s.SetAttr("href", domain+href)
		}
	})
}

// GetTagIDs 获取标签ids
func GetTagIDs(s string, tags []string) []uint {
	ids := []uint{}
	for _, v := range tags {
		if _, ok := model.Tags[v]; !ok {
			log.Fatal("tag不存在：", v)
		}
		ids = append(ids, model.Tags[v].ID)
	}

A:
	for _, v := range model.Tags {
		if strings.Contains(s, v.Name) {
			for _, tagName := range tags {
				if v.Name == tagName {
					continue A
				}
			}
			ids = append(ids, v.ID)
		}
	}
	return ids
}

// StrToTime 字符串转时间戳
func StrToTime(layout, val string) (t time.Time) {
	var err error
	loc, err := time.LoadLocation("Asia/Chongqing")
	if val != "" {
		t, err = time.ParseInLocation(layout, val, loc)
		if err != nil {
			log.Error("时间转换失败：", err.Error())
		}
	}
	return t
}

// GetCookies 获取cookie字符串
func GetCookies(url, tag string) ([]*http.Cookie, error) {
	var err error
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	chromeClient := client.New()
	chromeURL := client.URL(config.GetConfig("").ChromeURL)
	chromeURL(chromeClient)
	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithTargets(chromeClient.WatchPageTargets(ctxt)))
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	var cookies []*http.Cookie

	// run task list
	err = c.Run(ctxt, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(tag, chromedp.ByQuery),
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
