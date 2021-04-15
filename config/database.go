package config

type DatabaseConfiguration struct {
	Driver   string
	User     string
	Name     string
	Password string
	Port     int
	SSL      string
}
