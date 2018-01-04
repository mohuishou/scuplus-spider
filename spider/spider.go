package spider

import (
	"strings"

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

// GetTag 获取标签
func GetTag(s string, tags []string) []string {
	for _, v := range Tags {
		if strings.Contains(s, v) {
			tags = append(tags, v)
		}
	}
	return tags
}
