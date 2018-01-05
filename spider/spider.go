package spider

import (
	"strings"

	"github.com/mohuishou/scuplus-spider/log"

	"github.com/mohuishou/scuplus-spider/model"

	"github.com/PuerkitoBio/goquery"
)

// LinkHandle 替换img/a标签链接
func LinkHandle(contentDom *goquery.Selection, domain string) {
	contentDom.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exist := s.Attr("src"); exist {
			s.SetAttr("src", domain+src)
		}
	})
	contentDom.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exist := s.Attr("href"); exist {
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
