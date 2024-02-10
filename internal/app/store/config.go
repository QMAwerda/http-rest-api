package store

type Config struct {
	DatabaseURL string `toml:"database_url"` // строка подключения к нашей БД
}

func NewConfig() *Config {
	return &Config{}
}
