package redis

import (
	"time"

	"github.com/sevenNt/ares/environ"
)

var (
	conf *RedisConfig
)

type RedisConfig struct {
	environ.Base
	DialTimeout  time.Duration  `conf:"timeout.dial" flag:"dt" order:"conf,flag" default:"2s"`
	ReadTimeout  time.Duration  `conf:"timeout.read" default:"1s"`
	WriteTimeout time.Duration  `conf:"timeout.write" default:"1s"`
	IdleTimeout  time.Duration  `conf:"timeout.idle" default:"2s"`
	DB           int            `conf:"db" default:"1"`
	MaxIdle      int            `conf:"max.idle" default:"10"`
	MaxActive    int            `conf:"max.active" default:"50"`
	Wait         bool           `conf:"wait"`
	LogMode      bool           `env:"LOG_MODE"`
	Slaves       []string       `conf:"slaves"`
	Nodes        []int          `conf:"nodes"`
	Maps         map[string]int `conf:"maps"`
}

func Init() {
	if err := environ.Apply(&conf, environ.WithConfPrefix("redix")); err != nil {
		panic(err)
	}

}
