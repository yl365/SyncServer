// config.go
package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	LogLevel     string
	Listen       string
	ReqFreqLimit uint16

	RedisServer string
	RedisPasswd string
	RedisDB     uint16
	RedisKey    string

	UserRedisServer string
	UserRedisPasswd string
	UserRedisDB     uint16
	UserRedisKey    string

	UserMaxGrp uint16
	GrpMaxItem uint16
	SidTimeOut uint32
}

var g_conf Config

func LoadConfig() {
	if _, err := toml.DecodeFile("./config.toml", &g_conf); err != nil {
		Fatal("LoadConfig err=%s", err)
	}

	Info("LoadConfig g_conf=\n%+v", g_conf)
}
