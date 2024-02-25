package apiserver

type Config struct {
	BindAddr    string `toml:"bind_addr"` // тут используется обратный апостроф (grave accent)
	LogLevel    string `toml:"log_level"`
	DatabaseURL string `toml:"database_url"`
	SessionKey  string `toml:"session_key"`
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "debug",
	}
}
