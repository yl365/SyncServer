package main

import (
	"encoding/json"
	"fmt"
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

func Login(req string) string {
	defer profiler("Login", time.Now().UnixNano())

	var reqStruct Login_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(req)\"}"
	}

	respStr := ""
	sid := DoLogin(reqStruct)
	if sid > 0 {
		respStr = fmt.Sprintf("{\"Code\":0,\"Msg\":\"suc\",\"Sid\":%d}", sid)
	} else {
		respStr = fmt.Sprintf("{\"Code\":-1,\"Msg\":\"fail\"}")
	}

	return respStr
}

func Logout(req string) string {
	defer profiler("Logout", time.Now().UnixNano())

	var reqStruct Logout_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(req)\"}"
	}

	DoLogout(reqStruct.Sid)
	respStr := ("{\"Code\":0,\"Msg\":\"suc\"}")
	return respStr
}

func AllGrp(req string) string {
	defer profiler("AllGrp", time.Now().UnixNano())

	var reqStruct AllGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(req)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	resp, err := GetAllGrp(Uname)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(resp)\"}"
	}

	return string(respBody)
}

func CreateGrp(req string) string {
	defer profiler("CreateGrp", time.Now().UnixNano())

	var reqStruct CreateGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	//check Grp NUM
	if uint16(len(AllGrp.Grps)) >= g_conf.UserMaxGrp {
		return "{\"Code\":-1,\"Msg\":\"More than the largest number of groups\"}"
	}

	//check GrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.GrpName, V.GrpName) {
			return "{\"Code\":-1,\"Msg\":\"already existed\"}"
		}
	}

	storeData := &StoreFormat{Order: len(AllGrp.Grps), GrpID: uint64(time.Now().UnixNano() / 1000000), GrpName: reqStruct.GrpName, GrpVer: 0, Items: []string{}}
	err = SetGrpData(Uname, storeData.GrpID, storeData)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	resp := &CreateGrp_resp{Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail\"}"
	}

	return string(respBody)
}

func DeleteGrp(req string) string {
	defer profiler("DeleteGrp", time.Now().UnixNano())

	var reqStruct DeleteGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	err = DelGrp(Uname, reqStruct.GrpIDs)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	return "{\"Code\":0,\"Msg\":\"suc\"}"
}

func RenameGrp(req string) string {
	defer profiler("RenameGrp", time.Now().UnixNano())

	var reqStruct RenameGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	//check NewGrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.NewGrpName, V.GrpName) {
			return "{\"Code\":-1,\"Msg\":\"already existed\"}"
		}
	}

	storeData.GrpName = reqStruct.NewGrpName
	//storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	resp := &RenameGrp_resp{Code: 0, Msg: "suc", GrpID: reqStruct.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail\"}"
	}

	return string(respBody)
}

func ChangeGrpOrder(req string) string {
	defer profiler("ChangeOrderGrp", time.Now().UnixNano())

	var reqStruct ChangeGrpOrder_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	for i, GrpID := range reqStruct.GrpOrder {

		storeData, err := GetGrpData(Uname, GrpID)
		if err != nil {
			return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
		}

		storeData.Order = i
		err = SetGrpData(Uname, GrpID, storeData)
		if err != nil {
			return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
		}
	}

	return "{\"Code\":0,\"Msg\":\"suc\"}"
}

func Upload(req string) string {
	defer profiler("Upload", time.Now().UnixNano())

	var reqStruct Upload_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	bReturnItems := false
	if reqStruct.GrpVer != storeData.GrpVer {
		bReturnItems = true
	}

	switch reqStruct.Action {
	case 0:
		if uint16(len(storeData.Items))+uint16(len(reqStruct.Items)) > g_conf.GrpMaxItem {
			return "{\"Code\":-1,\"Msg\":\"More than the largest number of item\"}"
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
			return "{\"Code\":-1,\"Msg\":\"More than the largest number of item\"}"
		}
		storeData.Items = reqStruct.Items
	}
	storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
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
		return "{\"Code\":-1,\"Msg\":\"fail\"}"
	}

	return string(respBody)
}

func Download(req string) string {
	defer profiler("Download", time.Now().UnixNano())
	var reqStruct Download_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(data parse)\"}"
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return "{\"Code\":-1,\"Msg\":\"fail(Sid)\"}"
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail(db)\"}"
	}

	if reqStruct.GrpVer == storeData.GrpVer {
		respStr := fmt.Sprintf("{\"Code\":1,\"Msg\":\"isLatest\",\"GrpID\":%d, \"GrpVer\":%d}",
			storeData.GrpID, storeData.GrpVer)
		return respStr
	}

	resp := &Download_resp{Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	resp.Items = storeData.Items
	respBody, err := json.Marshal(resp)
	if err != nil {
		return "{\"Code\":-1,\"Msg\":\"fail\"}"
	}
	return string(respBody)
}
