package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type DataBase struct {
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Host     string `env:"PG_CONTAINER"`
	DB       string `env:"DB"`
}

type Config struct {
	DBPath []string
}

func ConfigLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	shardCount, err := strconv.Atoi(os.Getenv("SHARD_COUNT"))
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("DB")

	if err != nil {
		panic(err)
	}

	cfg := &Config{DBPath: make([]string, shardCount)}

	for i := 0; i < shardCount; i++ {
		container := os.Getenv(fmt.Sprintf("PG_SHARD_%d", i))
		cfg.DBPath[i] = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, container, db)
	}

	return cfg
}
