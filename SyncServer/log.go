package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LvlFatal = iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
)

/*
	一个log模块封装

	特性:
		日志级别支持
		日志路径/前缀设置: path/prefixmm-dd.HH.log
		每小时自动生成一个文件
		输出文件名行号, 方便排错
*/

var g_level int = LvlInfo
var g_path string = "./log"
var g_prefix string = "log" //filename: path/prefixmm-dd.HH.log
var g_lastReNewLog int64    //last renew log file time
var g_logfile *os.File = nil
var g_logLock sync.Mutex

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func SetLog(path, prefix string) {

	if len(path) > 0 {
		g_path = path
	}
	if len(prefix) > 0 {
		g_prefix = prefix
	}

	g_lastReNewLog = (time.Now().Unix() / 3600) * 3600

	filePathName := g_path + "/" + g_prefix + time.Now().Format("01-02.15") + ".log"

	var err error
	g_logfile, err = os.OpenFile(filePathName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("create log file err, exit! err =", err.Error(), filePathName)
		os.Exit(-1)
	}
}

func SetLogLevel(level string) {

	level = strings.ToLower(level)
	switch level {
	case "fatal":
		g_level = LvlFatal
	case "error":
		g_level = LvlError
	case "warn":
		g_level = LvlWarn
	case "info":
		g_level = LvlInfo
	case "debug":
		g_level = LvlDebug
	}
}

func reNewLog() {

	now := time.Now().Unix()

	if now-g_lastReNewLog < 3600 {
		return
	}

	if now-g_lastReNewLog >= 3600 {

		g_logfile.Close()

		filePathName := g_path + "/" + g_prefix + time.Now().Format("01-02.15") + ".log"

		var err error
		g_logfile, err = os.OpenFile(filePathName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			println("create log file err, exit! err =", err.Error())
			os.Exit(-1)
		}
		g_lastReNewLog = (now / 3600) * 3600
	}
}

func getTimeCodeLine() (string, string) {

	codeline := "(??:??)"
	if g_level >= LvlDebug {
		_, file, line, ok := runtime.Caller(2) //很慢,高级别就不打印行号
		if !ok {
			file = "???"
			line = 0
		}

		short := ""
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		codeline = fmt.Sprintf(" (%s:%d)", short, line)
	}

	now := time.Now()
	return fmt.Sprintf("%02d-%02d %02d:%02d:%02d.%03d", now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000), codeline
}

func Debug(format string, v ...interface{}) {

	if g_level < LvlDebug {
		return
	}

	g_logLock.Lock()
	reNewLog()

	time, line := getTimeCodeLine()
	fmt.Fprintf(g_logfile, time+" [D] "+format+line+"\n\n", v...)
	g_logLock.Unlock()
}

func Info(format string, v ...interface{}) {

	if g_level < LvlInfo {
		return
	}

	g_logLock.Lock()
	reNewLog()

	time, line := getTimeCodeLine()
	fmt.Fprintf(g_logfile, time+" [I] "+format+line+"\n\n", v...)
	g_logLock.Unlock()
}

func Warn(format string, v ...interface{}) {

	if g_level < LvlWarn {
		return
	}

	g_logLock.Lock()
	reNewLog()

	time, line := getTimeCodeLine()
	fmt.Fprintf(g_logfile, time+" [W] "+format+line+"\n\n", v...)
	g_logLock.Unlock()
}
func Error(format string, v ...interface{}) {

	if g_level < LvlError {
		return
	}

	g_logLock.Lock()
	reNewLog()

	time, line := getTimeCodeLine()
	fmt.Fprintf(g_logfile, time+" [E] "+format+line+"\n\n", v...)
	g_logLock.Unlock()
}
func Fatal(format string, v ...interface{}) {

	if g_level < LvlFatal {
		return
	}

	g_logLock.Lock()
	reNewLog()

	time, line := getTimeCodeLine()
	fmt.Fprintf(g_logfile, time+" [F] "+format+line+"\n\n", v...)
	g_logLock.Unlock()
	os.Exit(-1)
}
