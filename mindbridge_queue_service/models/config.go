package models

type CacheConfig struct {
	Host     string // CACHE_HOST
	PoolSize int    // CACHE_POOLSIZE
}

type LogConfig struct {
	LogFile  string
	LogLevel string
}

type HttpConfig struct {
	HostPort string
	HostCert string
	HostKey  string
}

type ServerConfig struct {
	OsUser               string
	EmailTemplatePath    string
	ServerHost           string
	SipServerHost        string
	SipServerHostPrivate string
	ServerProtocol       string
	ServerDomain         string
	ClientProfile        string
	Clip                 string
	ScheduledReportHour  string
}

type TimezoneConfig struct {
	ServerTimezone string
	DBTimezone     string
}

// Config - configuration object
type Config struct {
	Cache      CacheConfig
	Log        LogConfig
	HttpConfig HttpConfig
	Server     ServerConfig
	Timezone   TimezoneConfig
}
