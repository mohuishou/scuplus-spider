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
	model.DB().Where("category = ?", "scuinfo").Table("details").Count(&total)
	log.Println("total:", total)
	for i := 0; i < total/per; i++ {
		waitgroup.Add(1)
		go func(i int) {
			details := []model.Detail{}
			model.DB().Where("category = ?", "scuinfo").Offset(i * per).Limit(per).Select([]string{"id,created_at"}).Find(&details)
			for _, d := range details {
				// c := d.CreatedAt.String()
				u := d.CreatedAt.Unix()
				CreatedAt := time.Unix(u-1516416, u-1516416)
				model.DB().Model(&d).Update("created_at", CreatedAt)
				// log.Println(u, CreatedAt.Unix(), c, CreatedAt.String())
				// waitgroup.Done()
			}
			log.Printf("第%d批: %d ----> %d", i+1, details[0].ID, details[len(details)-1].ID)
			waitgroup.Done()

		}(i)

	}
	waitgroup.Wait()

}
