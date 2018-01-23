package main

import (
	"log"
	"sync"
	"time"

	"github.com/mohuishou/scuplus-spider/model"
)

func main() {
	var waitgroup sync.WaitGroup
	total := 0
	per := 100
	// model.DB().LogMode(true)
	start, _ := time.Parse("2006-01-02 15:04", "2018-01-05 00:00")
	end, _ := time.Parse("2006-01-02 15:04", "2018-01-05 23:59")
	scope := model.DB().Where("category = ? and created_at > ? and created_at < ? ", "青春川大", start, end)
	scope.Table("details").Count(&total)
	log.Println("total:", total, start, end)
	for i := 0; i < total/per+1; i++ {
		waitgroup.Add(1)
		go func(i int) {
			details := []model.Detail{}
			scope.Offset(i * per).Limit(per).Select([]string{"id,created_at,updated_at"}).Find(&details)
			for _, d := range details {
				CreatedAt := time.Unix(int64(d.ID), int64(d.ID))
				model.DB().Model(&d).Update("created_at", CreatedAt)
			}
			log.Printf("第%d批: %d ----> %d", i+1, details[0].ID, details[len(details)-1].ID)
			waitgroup.Done()

		}(i)

	}
	waitgroup.Wait()

}
