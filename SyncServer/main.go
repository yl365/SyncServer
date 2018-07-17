// StockProxy project main.go
package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type HttpHandler struct{}

var redisPool *redis.Pool
var UserRedisPool *redis.Pool

var staticHandler http.Handler = http.FileServer(http.Dir("./www/"))

func init() {
	rand.Seed(time.Now().Unix())
	SetLog("debug", "./log", "Log-")
	confinit()

	redisPool = newPool(g_conf.RedisServer, g_conf.RedisPasswd, int(g_conf.RedisDB))
	UserRedisPool = newPool(g_conf.UserRedisServer, g_conf.UserRedisPasswd, int(g_conf.UserRedisDB))
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	Info("ServeHTTP RemoteAddr=%s, r.URL.Path=%s, Form=%+v", r.RemoteAddr, r.URL.Path, r.Form)

	if IsLimit(r.RemoteAddr, g_conf.ReqFreqLimit) {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(ReqFreqLimit)\"}"))
		return
	}

	switch {
	case strings.EqualFold(r.URL.Path, "/api/v1/Login"):
		Login(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/Logout"):
		Logout(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/AllGrp"):
		AllGrp(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/CreateGrp"):
		CreateGrp(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/DeleteGrp"):
		DeleteGrp(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/RenameGrp"):
		RenameGrp(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/ChangeGrpOrder"):
		ChangeGrpOrder(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/Upload"):
		Upload(w, r)
	case strings.EqualFold(r.URL.Path, "/api/v1/Download"):
		Download(w, r)

	default:
		staticHandler.ServeHTTP(w, r)
	}
}

func main() {
	Info("\n\n\n##################### start...")
	go crontab()

	server := &http.Server{
		Addr:              ":9999",
		Handler:           &HttpHandler{},
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    2048,
	}

	err := server.ListenAndServe()
	if err != nil {
		Fatal("ListenAndServe: ", err)
	}
}

func crontab() {

	ticker1M := time.NewTicker(60 * time.Second)
	ticker5M := time.NewTicker(300 * time.Second)
	ticker1H := time.NewTicker(3600 * time.Second)
	for {
		select {
		case <-ticker1M.C:

		case <-ticker5M.C:

		case time := <-ticker1H.C:
			if time.Hour() == 3 { //每天凌晨3点清理
				ClearSID()
			}
		}
	}
	return
}
