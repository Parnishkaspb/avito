package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Server  ServerConfig    `yaml:"server"`
	Postgre PostreSQLConfig `yaml:"postgresql"`
	JWT     JWTConfig       `yaml:"jwt"`
}

type ServerConfig struct {
	Port    int           `yaml:"port"`
	Host    string        `yaml:"host"`
	Timeout time.Duration `yaml:"timeout"`
}

type PostreSQLConfig struct {
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	DB       string `yaml:"db"   env-required:"true"`
	SSLMode  string `yaml:"sslmode"  env-default:"disable"`
}

type JWTConfig struct {
	Secret string `yaml:"jwt_secret" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	fmt.Println(path)
	if path == "" {
		return &Config{}
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config path does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = ""
	}

	return res
}
