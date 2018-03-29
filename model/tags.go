package model

// Tag 标签
type Tag struct {
	Model
	Name    string   `gorm:"unique_index"`
	Details []Detail `gorm:"many2many:detail_tags"`
}

// tagNames 需要初始化捕获的标签
var tagNames = []string{
	"四川大学新闻网", "青春川大", "教务处", "学工部", "社团联", "研究生院", "就业网",
	"社团活动", "本科教育", "研究生教育", "科研动态", "院内动态", "院内通知", "会议通知", "院系风采", "团情快讯", "热门", "最近", "川大在线", "新闻", "公告", "比赛", "挑战杯", "凤凰展翅", "宣讲会", "招聘信息", "招生",
	"材料科学与工程学院", "电气信息学院", "电子信息学院", "法学院", "高分子科学与工程学院", "公共管理学院", "华西公共卫生学院", "华西口腔医学院", "华西临床医学院", "华西药学院", "化学学院", "化学工程学院", "华西基础医学与法医学院", "计算机学院", "建筑与环境学院", "经济学院", "匹兹堡学院", "历史文化学院", "轻纺与食品学院", "软件学院", "商学院", "生命科学学院", "数学学院", "水利水电学院", "外国语学院", "文学与新闻学院", "物理科学与技术学院", "艺术学院", "制造科学与工程学院",
}

// Tags 所有的标签
var Tags map[string]Tag

// 初始化获取所有标签，默认标签不存在则新建
func init() {
	if Tags == nil {
		Tags = map[string]Tag{}
		allTags := []Tag{}
		DB().Find(&allTags)

		for _, v := range allTags {
			Tags[v.Name] = v
		}
	}

	for _, name := range tagNames {
		if _, ok := Tags[name]; !ok {
			tmpTag := Tag{Name: name}
			DB().Create(&tmpTag)
			Tags[name] = tmpTag
		}
	}
}
