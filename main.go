// StockProxy project main.go
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/gomodule/redigo/redis"
)

var redisPool *redis.Pool
var UserRedisPool *redis.Pool

var staticHandler http.Handler = http.FileServer(http.Dir("./www/"))

func init() {
	rand.Seed(time.Now().Unix())
	SetLog("./log", "Log-")
	LoadConfig()
	SetLogLevel(g_conf.LogLevel)

	redisPool = newPool(g_conf.RedisServer, g_conf.RedisPasswd, int(g_conf.RedisDB))
	UserRedisPool = newPool(g_conf.UserRedisServer, g_conf.UserRedisPasswd, int(g_conf.UserRedisDB))
}

func ParseCmd(req string) string {

	respStr := ""
	Data := DataPackage{}
	err := json.Unmarshal([]byte(req), &Data)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	DataType := strings.ToLower(Data.Type)
	switch DataType {
	case "login":
		respStr = Login(Data.Tid, req)
	case "logout":
		respStr = Logout(Data.Tid, req)
	case "allgrp":
		respStr = AllGrp(Data.Tid, req)
	case "creategrp":
		respStr = CreateGrp(Data.Tid, req)
	case "deletegrp":
		respStr = DeleteGrp(Data.Tid, req)
	case "renamegrp":
		respStr = RenameGrp(Data.Tid, req)
	case "changegrporder":
		respStr = ChangeGrpOrder(Data.Tid, req)
	case "upload":
		respStr = Upload(Data.Tid, req)
	case "download":
		respStr = Download(Data.Tid, req)
	default:
		respStr = fmt.Sprintf("{\"Tid\":%u,\"Code\":-1,\"Msg\":\"fail(Invalid request)\"}", Data.Tid)
	}

	return respStr
}

func WsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		Info("WsHandlerFunc RemoteAddr=%s, r.URL.Path=%s, ws.UpgradeHTTP err=%+v", r.RemoteAddr, r.URL.Path, err.Error())
		return
	}

	go func() {
		defer conn.Close()

		for {
			req := ""
			respStr := ""

			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				Info("WsHandlerFunc wsutil.ReadClientData err=%s", err.Error())
				break
			}

			if g_conf.ReqFreqLimit > 0 && IsLimit(r.RemoteAddr, g_conf.ReqFreqLimit) {
				respStr = "{\"Code\":-1,\"Msg\":\"fail(ReqFreqLimit)\"}"
			} else {
				req = string(msg)
				respStr = ParseCmd(req)
			}

			err = wsutil.WriteServerMessage(conn, op, []byte(respStr))
			if err != nil {
				Info("WsHandlerFunc wsutil.WriteServerMessage err=%s", err.Error())
				break
			}

			Info("WsHandlerFunc RemoteAddr=%s, r.URL.Path=%s, req=%s, resp=%s", r.RemoteAddr, r.URL.Path, req, respStr)
		}
	}()
}

func ApiHandlerFunc(w http.ResponseWriter, r *http.Request) {

	for {
		req := ""
		respStr := ""

		if g_conf.ReqFreqLimit > 0 && IsLimit(r.RemoteAddr, g_conf.ReqFreqLimit) {
			respStr = "{\"Code\":-1,\"Msg\":\"fail(ReqFreqLimit)\"}"
			break
		}

		req = r.Form.Get("req")
		req, err := url.QueryUnescape(req)
		if err != nil {
			respStr = "{\"Code\":-1,\"Msg\":\"fail(req data)\"}"
			break
		}

		respStr = ParseCmd(req)
		break
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(respStr))
	Info("ApiHandlerFunc RemoteAddr=%s, r.URL.Path=%s, req=%s, resp=%s", r.RemoteAddr, r.URL.Path, req, respStr)
}

func HttpHandlerFunc(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	Info("HttpHandlerFunc RemoteAddr=%s, r.URL.Path=%s, Form=%+v", r.RemoteAddr, r.URL.Path, r.Form)

	switch {
	case strings.EqualFold(r.URL.Path, "/api"):
		ApiHandlerFunc(w, r)

	case strings.EqualFold(r.URL.Path, "/ws"):
		WsHandlerFunc(w, r)

	default:
		staticHandler.ServeHTTP(w, r)
	}
}

func main() {
	Info("##################### start...")
	go crontab()

	server := &http.Server{
		Addr:              g_conf.Listen,
		Handler:           http.HandlerFunc(HttpHandlerFunc),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    2048,
	}

	err := server.ListenAndServe()
	if err != nil {
		Fatal("ListenAndServe: %s", err)
	}
}

func crontab() {

	if g_conf.EnableProfile {
		cpuf, _ := os.Create("cpu_profile")
		pprof.StartCPUProfile(cpuf)
	}

	StopProfileFun := func() {
		if g_conf.EnableProfile {
			pprof.StopCPUProfile()
			memf, _ := os.Create("mem_profile")
			pprof.WriteHeapProfile(memf)
			memf.Close()
		}
	}

	ticker1M := time.NewTicker(60 * time.Second)
	ticker5M := time.NewTicker(300 * time.Second)
	ticker1H := time.NewTicker(3600 * time.Second)
	for {
		select {
		case <-ticker1M.C:
			StopProfileFun()

		case <-ticker5M.C:

		case time := <-ticker1H.C:
			if time.Hour() == 3 { //每天凌晨3点清理
				ClearSID()
			}
		}
	}
	return
}
