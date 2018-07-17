package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

// 会话ID管理
type SID struct {
	Uname     string
	LoginTime uint32 //登录时间
	LastAlive uint32 //最后活跃时间
}

var g_SID map[uint64]SID
var g_SIDlock sync.Mutex
var g_count uint64

func init() {
	g_SIDlock.Lock()
	if g_SID == nil {
		g_SID = make(map[uint64]SID, 10000)
	}
	g_SIDlock.Unlock()
}

func DoLogin(login Login_req) uint64 {

	redisConn := UserRedisPool.Get()
	defer redisConn.Close()

	login.Uname = strings.ToLower(login.Uname)
	UserRedisKey := fmt.Sprintf(g_conf.UserRedisKey, login.Uname)
	passwd, err := redis.String(redisConn.Do("GET", UserRedisKey))
	if err != nil || passwd != login.Passwd {
		return 0
	}

	nowTS := uint32(time.Now().Unix())
	g_SIDlock.Lock()

	g_count++
	sid := uint64(nowTS-1500000000)*1000000000 + uint64(g_count%100000)*10000 + uint64(rand.Intn(9999))
	newSID := SID{Uname: login.Uname, LoginTime: nowTS, LastAlive: nowTS}
	g_SID[sid] = newSID

	g_SIDlock.Unlock()
	Info("DoLogin: Uname=%s <--> Sid=%d ", login.Uname, sid)
	return sid
}

func DoLogout(sid uint64) {

	if g_SID == nil || sid == 0 {
		return
	}

	g_SIDlock.Lock()
	delete(g_SID, sid)
	g_SIDlock.Unlock()
}
func CheckSid(sid uint64) (string, error) {

	if sid == 0 {
		return "", errors.New("Invalid session identifier")
	}
	if g_SID == nil {
		return "", errors.New("Invalid session identifier")
	}

	nowTS := uint32(time.Now().Unix())
	g_SIDlock.Lock()
	defer g_SIDlock.Unlock()
	oldSID, have := g_SID[sid]

	if have {
		if nowTS-oldSID.LastAlive < g_conf.SidTimeOut {
			Info("CheckSid: Sid=%d <--> Uname=%s", sid, oldSID.Uname)
			oldSID.LastAlive = nowTS
			g_SID[sid] = oldSID
			return oldSID.Uname, nil
		} else {
			Info("CheckSid: Sid=%d <--> Uname=%s [TimeOut]", sid, oldSID.Uname)
			delete(g_SID, sid)
		}
	}

	Info("CheckSid: Sid=%d [Invalid]", sid)
	return "", errors.New("Invalid session identifier")
}

func ClearSID() {

	if g_SID == nil {
		return
	}

	g_SIDlock.Lock()
	nowTS := uint32(time.Now().Unix())
	Info("ClearSID before g_SID NUM=%d", len(g_SID))
	for sid, oldSID := range g_SID {
		if nowTS-oldSID.LastAlive >= g_conf.SidTimeOut {
			delete(g_SID, sid)
		}
	}
	Info("ClearSID after g_SID NUM=%d", len(g_SID))
	g_SIDlock.Unlock()
}
