// config.go
package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
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

func confinit() {
	if _, err := toml.DecodeFile("./config.toml", &g_conf); err != nil {
		Fatal("confinit err=%s", err)
	}

	Info("confinit g_conf=\n%+v", g_conf)
}
