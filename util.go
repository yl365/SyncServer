package main

import (
	"strings"
	"time"
)

func profiler(funcName string, start int64) {
	Debug("### FUNC[%s] cost=%.03fms",
		funcName, (float64(time.Now().UnixNano())-float64(start))/1000000.0)
}

func getUInt8(buf []byte) uint8 {
	return uint8(buf[0])
}
func getUInt16(buf []byte) uint16 {
	t := uint16(buf[0])
	t = t | uint16(buf[1])<<8
	return t
}
func getUInt32(buf []byte) uint32 {
	t := uint32(buf[0])
	t = t | uint32(buf[1])<<8
	t = t | uint32(buf[2])<<16
	t = t | uint32(buf[3])<<24
	return t
}
func getUInt64(buf []byte) uint64 {
	t := uint64(buf[0])
	t = t | uint64(buf[1])<<8
	t = t | uint64(buf[2])<<16
	t = t | uint64(buf[3])<<24
	t = t | uint64(buf[1])<<32
	t = t | uint64(buf[2])<<40
	t = t | uint64(buf[3])<<48
	t = t | uint64(buf[3])<<56
	return t
}

func UrlArray(ignorePrefix, url string) []string {

	cutPrefix := url[len(ignorePrefix):]
	ParamArray := strings.Split(cutPrefix, "/")

	return ParamArray
}

func UrlMap(ignorePrefix, url string) map[string]string {

	cutPrefix := url[len(ignorePrefix):]
	ParamArray := strings.Split(cutPrefix, "/")

	ArrayLen := len(ParamArray)
	if ArrayLen%2 != 0 { //如果最后一个参数只有key,需要补0
		ParamArray[ArrayLen] = "0"
	}

	var ParamMap map[string]string
	for i := 0; i < len(ParamArray); i += 2 {
		ParamMap[ParamArray[i]] = ParamArray[i+1]
	}

	return ParamMap
}

const (
	MINMASK   = 0X3F
	HOURMASK  = 0X7C0
	DAYMASK   = 0XF800
	MONTHMASK = 0XF0000
	YEARMASK  = 0XFFF00000
)

func timeConv(rq uint32) uint64 {

	rq64 := uint64(rq)
	min := rq64 & MINMASK
	hour := (rq64 & HOURMASK) >> 6
	day := (rq64 & DAYMASK) >> 11
	month := (rq64 & MONTHMASK) >> 16
	year := (rq64 & YEARMASK) >> 20

	return year*100000000 + month*1000000 + day*10000 + hour*100 + min
}

// 截取两个字符串中间的字符串 <YZMPT><IMSI>1234567890</IMSI><CUST>xxxx-xxxx</CUST></YZMPT>
func getPairText(text, start, end string) string {

	UpperText := strings.ToUpper(text)
	s := strings.Index(UpperText, strings.ToUpper(start))
	if s == -1 {
		return ""
	}

	e := strings.Index(UpperText[s:], strings.ToUpper(end))
	if e == -1 {
		return ""
	}

	return text[s+len(start) : s+e]
}
