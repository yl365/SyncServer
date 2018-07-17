package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

//请求频率限制:
//		同一个IP每5分钟请求超过N次, 超限禁用5分钟

type min5 struct {
	pos uint16
	num uint16
}

var g_limitlock sync.Mutex
var g_limit map[uint32]min5
var g_limitClearTime int64

func init() {
	g_limit = make(map[uint32]min5, 100)
}

func Ip2long(ipstr string) (ip uint32) {

	ips := strings.Split(ipstr, ".")
	ip1, _ := strconv.Atoi(ips[0])
	ip2, _ := strconv.Atoi(ips[1])
	ip3, _ := strconv.Atoi(ips[2])
	ip4, _ := strconv.Atoi(ips[3])

	if ip1 > 255 || ip2 > 255 || ip3 > 255 || ip4 > 255 {
		return
	}

	ip += uint32(ip1 * 0x1000000)
	ip += uint32(ip2 * 0x10000)
	ip += uint32(ip3 * 0x100)
	ip += uint32(ip4)

	return
}

func Long2ip(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}

func IsLimit(RemoteAddr string, MAX uint16) bool {
	ret := false
	st := time.Now().Unix()
	pos := uint16(st / 300)

	ip := strings.Split(RemoteAddr, ":")[0]
	nIP := Ip2long(ip)

	g_limitlock.Lock()
	v, have := g_limit[nIP]
	if have {
		if v.pos != pos { //新的5分钟
			v.pos = pos
			v.num = 1
		} else {
			v.num++
			if v.num > MAX {
				ret = true
			}
		}
	} else {
		v.pos = pos
		v.num = 1
	}
	g_limit[nIP] = v

	//清理g_limit
	if st-g_limitClearTime >= 600 {
		for k, _ := range g_limit {
			delete(g_limit, k)
		}
		g_limitClearTime = st
	}

	g_limitlock.Unlock()

	return ret
}
