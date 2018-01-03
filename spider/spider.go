package spider

import (
	"github.com/PuerkitoBio/goquery"
)

// LinkHandle 替换图片链接
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
