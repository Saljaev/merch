package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type DataBase struct {
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Host     string `env:"PG_CONTAINER"`
	DB       string `env:"DB"`
}

type Config struct {
	DBPath      []string
	ShardNumber int
	JWTSecret   string
	Issuer      string
	TokenTTL    time.Duration
	CacheTTL    time.Duration
	SLI         time.Duration
	ADDR        string
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

	secret := os.Getenv("SECRET")
	issuer := os.Getenv("ISSUER")
	tokenTTL, err := parseDuration(os.Getenv("TOKEN_TTL"))
	if err != nil {
		panic(err)
	}

	cacheTTL, err := parseDuration(os.Getenv("CACHE_TTL"))
	if err != nil {
		panic(err)
	}

	sli, err := parseDuration(os.Getenv("SLI"))
	if err != nil {
		panic(err)
	}

	addr := os.Getenv("ADDR")

	cfg := &Config{
		DBPath:      make([]string, shardCount),
		ShardNumber: shardCount,
		JWTSecret:   secret,
		Issuer:      issuer,
		TokenTTL:    tokenTTL,
		CacheTTL:    cacheTTL,
		SLI:         sli,
		ADDR:        addr,
	}

	for i := 0; i < shardCount; i++ {
		container := os.Getenv(fmt.Sprintf("PG_SHARD_%d", i))
		cfg.DBPath[i] = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, container, db)
	}

	return cfg
}

func parseDuration(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration, nil
	}

	minutes, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ttl: %s", value)
	}

	return time.Duration(minutes) * time.Minute, nil
}
