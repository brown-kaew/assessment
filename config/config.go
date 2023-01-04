package config

import "os"

type Config struct {
	Port        string
	DatabaseUrl string
}

func New() Config {
	return Config{
		Port:        os.Getenv("PORT"),
		DatabaseUrl: os.Getenv("DATABASE_URL"),
	}
}
