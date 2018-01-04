package scuinfo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/config"
	"github.com/mohuishou/scuplus-spider/log"
)

var urls = map[string]string{
	"最近": "posts",
	"热门": "hot",
}

//Item scuinfo 单条数据
type Item struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Gender       int    `json:"gender"`
	Secret       int    `json:"secret"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	CommentCount int    `json:"comment_count"`
	Author       int    `json:"author"`
	UserID       int    `json:"user_id"`
	Date         int64  `json:"date"`
	More         int    `json:"more"`
}

// Data 返回值
type Data struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Items   []Item `json:"data"`
}

// 抓取数据
func spider(conf config.Spider) {
	if _, ok := urls[conf.Key]; !ok {
		log.Fatal("[E]: 不存在这个key")
	}
	url := fmt.Sprintf("http://scuinfo.com/api/%s?pageSize=15", urls[conf.Key])

	tryCount := 0

	c := colly.NewCollector()

	c.OnResponse(func(resp *colly.Response) {
		data := &Data{}
		err := json.Unmarshal(resp.Body, data)
		if err != nil {
			log.Error("获取数据错误,", err.Error())
		}
		if data.Code != 200 {
			log.Error("[E]: 获取数据错误")
		}

		// 处理数据
		for _, item := range data.Items {
			if conf.IsNew {
				//达到最大尝试次数，不再获取新的数据
				if tryCount > conf.MaxTryNum {
					return
				}

				if (time.Now().Unix() - item.Date) > int64(conf.Second) {
					tryCount++
					log.Info("数据已过期，将被丢弃")
					continue
				}
			}

			// 数据持久化
			detail := &model.Detail{
				Title:     item.Nickname,
				Content:   item.Content,
				Category:  "scuinfo",
				URL:       resp.Request.URL.String(),
				CreatedAt: item.Date,
			}

			detail.Create([]string{urls[conf.Key]})
		}

		// 发现新的页面
		fromID := data.Items[len(data.Items)-1].ID
		c.Visit(fmt.Sprintf("%s&&fromId=%d", url, fromID))
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Referer", "http://scuinfo.com")
	})

	c.Visit(url)
}
