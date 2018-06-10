package lecture

import (
	"strings"

	"log"

	"time"

	"github.com/gocolly/colly"
	"github.com/mohuishou/scuplus-spider/model"
	"github.com/mohuishou/scuplus-spider/spider"
)

const domain = "http://www.scu.edu.cn/index/xx/xskb/"
const max = 50

func Spider() {
	c := spider.NewCollector()

	// 是否继续获取
	count := 0

	// 获取下一页
	c.OnHTML("span.p_next.p_fun > a", func(e *colly.HTMLElement) {
		log.Println(domain + strings.Trim(e.Attr("href"), "xskb/"))
		if count > max {
			log.Println("lalalla")
			return
		}

		go c.Visit(domain + strings.Trim(e.Attr("href"), "xskb/"))
	})

	// 获取内容
	c.OnHTML(".kanban-List > li", func(e *colly.HTMLElement) {
		if count > max {
			return
		}

		data := model.Lecture{}

		// 获取时间
		timeStr := strings.TrimSpace(e.ChildText("p:nth-child(2) > span"))
		data.Time = timeStr
		arr := strings.Split(timeStr, "-")
		str := strings.Replace(arr[0], "：", ":", -1)
		str = strings.Replace(str, "上午", "", -1)
		str = strings.Replace(str, "下午", "", -1)
		t, err := time.Parse("2006年01月02日 15:04", str)
		if err != nil {
			log.Println("时间获取失败！")
			return
		}
		data.StartTime = t

		// 获取其他信息
		data.Title = strings.TrimSpace(e.ChildText("h3"))
		data.Address = strings.TrimSpace(e.ChildText("p:nth-child(3)"))
		data.Address = strings.Trim(data.Address, "地点：")
		data.Reporter = strings.TrimSpace(e.ChildText("p:nth-child(4)"))
		data.Reporter = strings.Trim(data.Reporter, "主讲人：")
		data.College = strings.TrimSpace(e.ChildText("p:nth-child(5)"))
		data.College = strings.Trim(data.College, "主办：")

		// 保存
		count++
		if err := model.DB().Create(&data).Error; err != nil {
			log.Println(err)
		}
	})

	c.Visit("http://www.scu.edu.cn/index/xx/xskb.htm")
	c.Wait()
}
