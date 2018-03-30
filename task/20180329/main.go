package main

import (
	"log"
	"sync"

	"github.com/mohuishou/scuplus-spider/model"
)

var category = []string{
	"青春川大", "教务处", "四川大学新闻网", "社团联", "学工部",
}

var w = sync.WaitGroup{}

func main() {
	for _, c := range category {
		w.Add(1)
		go func(v string) {
			t := model.Tags[v]
			log.Println(v, "开始转换", t)
			details := []model.Detail{}
			model.DB().Order("id desc").Where("category = ?", v).Select([]string{"id"}).Find(&details)
			for _, d := range details {
				dt := model.DetailTag{}
				err := model.DB().Where(&model.DetailTag{
					DetailID: d.ID,
					TagID:    t.ID,
				}).Find(&dt).Error
				if dt.TagID == 0 {
					err = model.DB().Create(&model.DetailTag{
						DetailID: d.ID,
						TagID:    t.ID,
					}).Error
					log.Println("转换成功：", d.ID)
				}
				if err != nil {
					log.Println(err)
				}
			}
			log.Println(v, "转换成功:", len(details))
			w.Done()
		}(c)
	}
	w.Wait()
}
