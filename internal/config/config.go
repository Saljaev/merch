package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
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
	ShardNumber int `yaml:"shard_number"`
	JWTSecret   string
	Issuer      string
	TokenTTL    time.Duration
	CacheTTL    time.Duration
	SLI         time.Duration `yaml:"sli"`
	ADDR        string
	MaxConn     int           `yaml:"max_conn"`
	MinConn     int           `yaml:"min_conn"`
	LifeConn    time.Duration `yaml:"life_conn"`
}

func ConfigLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var cfg Config

	readConfigYaml(&cfg)

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("DB")
	secret := os.Getenv("SECRET")
	issuer := os.Getenv("ISSUER")
	token_ttl, err := parseDuration(os.Getenv("TOKEN_TTL"))
	if err != nil {
		panic(err)
	}
	cache_ttl, err := parseDuration(os.Getenv("CACHE_TTL"))
	if err != nil {
		panic(err)
	}

	addr := os.Getenv("ADDR")

	cfg.DBPath = make([]string, cfg.ShardNumber)
	cfg.JWTSecret = secret
	cfg.Issuer = issuer
	cfg.ADDR = addr
	cfg.TokenTTL = token_ttl
	cfg.CacheTTL = cache_ttl

	for i := 0; i < cfg.ShardNumber; i++ {
		container := os.Getenv(fmt.Sprintf("PG_SHARD_%d", i))
		cfg.DBPath[i] = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, container, db)
	}

	return &cfg
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

func readConfigYaml(cfg *Config) {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		panic(err)
	}

}
