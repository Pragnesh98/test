package configs

import (
	"mindbridge_queue_service/models"
	"strings"

	"github.com/spf13/viper"
)

var conf *models.Config

// GetConfig - Function to get Config
func GetConfig() *models.Config {
	if conf != nil {
		return conf
	}
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cacheConf := models.CacheConfig{
		Host:     strings.TrimSpace(v.GetString("cache.host")),
		PoolSize: v.GetInt("cache.poolsize"),
	}

	logConf := models.LogConfig{
		LogFile:  strings.TrimSpace(v.GetString("log.file")),
		LogLevel: strings.TrimSpace(v.GetString("log.level")),
	}

	httpConf := models.HttpConfig{
		HostPort: strings.TrimSpace(v.GetString("http.host")),
	}

	conf = &models.Config{
		Cache:      cacheConf,
		Log:        logConf,
		HttpConfig: httpConf,
	}

	return conf
}
