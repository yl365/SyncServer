package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

func GetAllGrp(Uname string) (*AllGrp_resp, error) {

	resp := &AllGrp_resp{Code: -1, Msg: "fail", Grps: []GrpInfo{}}
	redisConn := redisPool.Get()
	defer redisConn.Close()

	RedisKey := fmt.Sprintf(g_conf.RedisKey, Uname)
	fieldValueMap, err := redis.StringMap(redisConn.Do("hgetall", RedisKey))
	if err != nil {
		return resp, err
	}

	var storeData StoreFormat
	var grpInfo GrpInfo
	for _, value := range fieldValueMap {

		err = json.Unmarshal([]byte(value), &storeData)
		if err == nil {
			grpInfo.Order = storeData.Order
			grpInfo.GrpID = storeData.GrpID
			grpInfo.GrpName = storeData.GrpName
			grpInfo.GrpVer = storeData.GrpVer
			resp.Grps = append(resp.Grps, grpInfo)
		}
	}
	sort.Sort(resp.Grps)

	resp.Code = 0
	resp.Msg = "suc"
	return resp, nil
}

func GetGrpData(Uname string, GrpID uint64) (*StoreFormat, error) {

	storeData := &StoreFormat{}
	redisConn := redisPool.Get()
	defer redisConn.Close()

	RedisKey := fmt.Sprintf(g_conf.RedisKey, Uname)
	Value, err := redis.String(redisConn.Do("hget", RedisKey, GrpID))
	if err != nil {
		return storeData, err
	}

	err = json.Unmarshal([]byte(Value), &storeData)
	if err != nil {
		return storeData, err
	}

	return storeData, nil
}

func SetGrpData(Uname string, GrpID uint64, storeData *StoreFormat) error {

	storeJson, err := json.Marshal(storeData)
	if err != nil {
		return err
	}

	redisConn := redisPool.Get()
	defer redisConn.Close()

	RedisKey := fmt.Sprintf(g_conf.RedisKey, Uname)
	_, err = redis.Int(redisConn.Do("hset", RedisKey, GrpID, storeJson))
	if err != nil {
		return err
	}

	return nil
}

func DelGrp(Uname string, GrpIDs []uint64) error {

	redisConn := redisPool.Get()
	defer redisConn.Close()

	RedisKey := fmt.Sprintf(g_conf.RedisKey, Uname)
	for _, ID := range GrpIDs {
		_, err := redis.Int(redisConn.Do("hdel", RedisKey, ID))
		if err != nil {
			return err
		}
	}

	return nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	defer profiler("Login", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json;charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}
	var reqStruct Login_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}

	respStr := ""
	sid := DoLogin(reqStruct)
	if sid > 0 {
		respStr = fmt.Sprintf("{\"Code\":0,\"Msg\":\"suc\",\"Sid\":%d}", sid)
	} else {
		respStr = fmt.Sprintf("{\"Code\":-1,\"Msg\":\"fail\"}")
	}
	w.Write([]byte(respStr))
	return
}

func Logout(w http.ResponseWriter, r *http.Request) {
	defer profiler("Logout", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json;charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}
	var reqStruct Logout_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}

	DoLogout(reqStruct.Sid)
	respStr := ("{\"Code\":0,\"Msg\":\"suc\"}")
	w.Write([]byte(respStr))
	return
}

func AllGrp(w http.ResponseWriter, r *http.Request) {
	defer profiler("AllGrp", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json;charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}
	var reqStruct AllGrp_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	resp, err := GetAllGrp(Uname)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(resp)\"}"))
		return
	}

	w.Write(respBody)
	return
}

func CreateGrp(w http.ResponseWriter, r *http.Request) {
	defer profiler("CreateGrp", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct CreateGrp_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	//check Grp NUM
	if uint16(len(AllGrp.Grps)) >= g_conf.UserMaxGrp {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"More than the largest number of groups\"}"))
		return
	}

	//check GrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.GrpName, V.GrpName) {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"already existed\"}"))
			return
		}
	}

	storeData := &StoreFormat{Order: len(AllGrp.Grps), GrpID: uint64(time.Now().UnixNano() / 1000000), GrpName: reqStruct.GrpName, GrpVer: 0, Items: []string{}}
	err = SetGrpData(Uname, storeData.GrpID, storeData)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	resp := &CreateGrp_resp{Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail\"}"))
		return
	}

	w.Write(respBody)
	return
}

func DeleteGrp(w http.ResponseWriter, r *http.Request) {
	defer profiler("DeleteGrp", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct DeleteGrp_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	err = DelGrp(Uname, reqStruct.GrpIDs)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	w.Write([]byte("{\"Code\":0,\"Msg\":\"suc\"}"))
	return
}

func RenameGrp(w http.ResponseWriter, r *http.Request) {
	defer profiler("RenameGrp", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct RenameGrp_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	//check NewGrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.NewGrpName, V.GrpName) {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"already existed\"}"))
			return
		}
	}

	storeData.GrpName = reqStruct.NewGrpName
	//storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	resp := &RenameGrp_resp{Code: 0, Msg: "suc", GrpID: reqStruct.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail\"}"))
		return
	}

	w.Write(respBody)
	return
}

func ChangeGrpOrder(w http.ResponseWriter, r *http.Request) {
	defer profiler("ChangeOrderGrp", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct ChangeGrpOrder_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	for i, GrpID := range reqStruct.GrpOrder {

		storeData, err := GetGrpData(Uname, GrpID)
		if err != nil {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
			return
		}

		storeData.Order = i
		err = SetGrpData(Uname, GrpID, storeData)
		if err != nil {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
			return
		}
	}

	w.Write([]byte("{\"Code\":0,\"Msg\":\"suc\"}"))
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	defer profiler("Upload", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct Upload_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	bReturnItems := false
	if reqStruct.GrpVer != storeData.GrpVer {
		bReturnItems = true
	}

	switch reqStruct.Action {
	case 0:
		if uint16(len(storeData.Items))+uint16(len(reqStruct.Items)) > g_conf.GrpMaxItem {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"More than the largest number of item\"}"))
			return
		}
		reqStruct.Items = append(reqStruct.Items, storeData.Items...)
		storeData.Items = reqStruct.Items
	case 1:
		tmpItems := storeData.Items
		storeData.Items = []string{}

		for _, stock := range tmpItems {
			bFind := false
			for _, curStock := range reqStruct.Items {
				if stock == curStock {
					bFind = true
					break
				}
			}
			if !bFind {
				storeData.Items = append(storeData.Items, stock)
			}
		}
	case 2:
		if uint16(len(reqStruct.Items)) > g_conf.GrpMaxItem {
			w.Write([]byte("{\"Code\":-1,\"Msg\":\"More than the largest number of item\"}"))
			return
		}
		storeData.Items = reqStruct.Items
	}
	storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	var respBody []byte
	if bReturnItems {
		resp := &Download_resp{Code: 2, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
		resp.Items = storeData.Items
		respBody, err = json.Marshal(resp)
	} else {
		resp := &Upload_resp{Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
		respBody, err = json.Marshal(resp)
	}

	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail\"}"))
		return
	}
	w.Write(respBody)
	return
}

func Download(w http.ResponseWriter, r *http.Request) {
	defer profiler("Download", time.Now().UnixNano())
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	req := r.Form.Get("req")
	req, err := url.QueryUnescape(req)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(req data)\"}"))
		return
	}
	var reqStruct Download_req
	err = json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"))
		return
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"))
		return
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail(db)\"}"))
		return
	}

	if reqStruct.GrpVer == storeData.GrpVer {
		respStr := fmt.Sprintf("{\"Code\":1,\"Msg\":\"isLatest\",\"GrpID\":%d, \"GrpVer\":%d}",
			storeData.GrpID, storeData.GrpVer)
		w.Write([]byte(respStr))
		return
	}

	resp := &Download_resp{Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	resp.Items = storeData.Items
	respBody, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte("{\"Code\":-1,\"Msg\":\"fail\"}"))
		return
	}
	w.Write(respBody)
	return
}
