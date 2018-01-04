package config

// Spider 爬虫设置
type Spider struct {
	MaxTryNum int      //最大尝试次数
	IsNew     bool     //是否获取最新的数据: true: 获取最新的数据, false: 获取所有的数据
	Key       string   //链接key
	Tags      []string //标签，会从标题或者是正文当中进行匹配
	Second    int      //获取距离当前时间多少秒之内的数据，仅在最新的数据时有效
}

// Mysql 配置
type Mysql struct {
	Host     string
	User     string
	Password string
	DB       string
	Port     string
}

// Config 对应config.yml文件的位置
type Config struct {
	Mysql
}

var config *Config

// GetConfig 获取config
func GetConfig() *Config {
	if config == nil {

	}
	return config
}
