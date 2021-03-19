package sq

import (
	"net/url"
	"strings"
)

type MysqlDataSource struct {
	User string `yaml:"user"`
	Password string `yaml:"password"`
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	DB string `yaml:"db"`
	// DefaultQuery
	// 	map[string]string{
	// 	"charset": "utf8",
	// 	"parseTime": "True",
	// 	"loc": "Local",
	// }
	Query map[string]string
}
func (config MysqlDataSource) String() (dataSourceName string) {
	configList := []string{
		config.User,
		":",
		config.Password,
		"@",
		"(",
		config.Host,
		":",
		config.Port,
		")",
		"/",
		config.DB,
		"?",
	}
	configList = append(configList)
	if config.Query == nil {
		config.Query = map[string]string{
				"charset": "utf8",
				"parseTime": "True",
				"loc": "Local",
			}
	}
	values := url.Values{}
	for key, value := range config.Query {
		values.Set(key, value)
	}
	dataSourceName = strings.Join(configList,"") + values.Encode()
	return
}
