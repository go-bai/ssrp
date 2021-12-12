package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Backends            []Backend
	TimeOut             int64
	HealthCheckInterval int64
}

type Backend struct {
	Port     string
	Host     string
	Urls     []string
	BackUp   string
	Upstream string
	TlsCert  string
	TlsKey   string
}

func (c *Config) Read() (err error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	allSettings := viper.AllSettings()
	backends := make([]Backend, 0)

	upstream := make(map[string][]string, 0)

	for key, value := range allSettings {
		switch key {
		case "timeout":
			c.TimeOut = value.(int64)
			continue
		case "healthcheck":
			c.HealthCheckInterval = value.(int64)
			continue
		default:
			if key != "backends" {
				// upstream
				urls := make([]string, 0)
				for _, v := range value.([]interface{}) {
					urls = append(urls, v.(string))
				}
				upstream[key] = urls
				continue
			}
		}

		for k, v := range value.(map[string]interface{}) {
			bk := v.(map[string]interface{})

			if bk["url"] == nil && bk["upstream"] == nil {
				return fmt.Errorf("url and upstream not defined for backend %s", key)
			}

			backend := Backend{
				Port: k,
			}
			if bk["host"] != nil {
				backend.Host = bk["host"].(string)
			}
			if bk["upstream"] != nil {
				backend.Upstream = bk["upstream"].(string)
			}
			if bk["backup"] != nil {
				backend.BackUp = bk["backup"].(string)
			}
			if bk["tls_cert"] != nil {
				backend.TlsCert = bk["tls_cert"].(string)
			}
			if bk["tls_key"] != nil {
				backend.TlsKey = bk["tls_key"].(string)
			}
			urls := bk["url"].([]interface{})
			for _, url := range urls {
				backend.Urls = append(backend.Urls, url.(string))
			}
			backends = append(backends, backend)
		}

	}
	for i, bk := range backends {
		if bk.Upstream != "" {
			backends[i].Urls = append(bk.Urls, upstream[bk.Upstream]...)
		}
	}
	c.Backends = backends
	return nil
}
