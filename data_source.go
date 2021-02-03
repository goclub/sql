package sq

import (
	"net/url"
	"strings"
)

type DataSourceName struct {
	DriverName string
	User string
	Password string
	Host string
	Port string
	DB string
	// 	map[string]string{
	// 	"charset": "utf8",
	// 	"parseTime": "True",
	// 	"loc": "Local",
	// }
	Query map[string]string
}
func (config DataSourceName) String() (dataSourceName string) {
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
