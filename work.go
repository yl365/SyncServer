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

func Login(Tid uint32, req string) string {
	defer profiler("Login", time.Now().UnixNano())

	var reqStruct Login_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(req)\"}", Tid)
	}

	respStr := ""
	sid := DoLogin(reqStruct)
	if sid > 0 {
		respStr = fmt.Sprintf("{\"Tid\":%d,\"Code\":0,\"Msg\":\"suc\",\"Sid\":%d}", Tid, sid)
	} else {
		respStr = fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail\"}", Tid)
	}

	return respStr
}

func Logout(Tid uint32, req string) string {
	defer profiler("Logout", time.Now().UnixNano())

	var reqStruct Logout_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(req)\"}", Tid)
	}

	DoLogout(reqStruct.Sid)
	return fmt.Sprintf("{\"Tid\":%d,\"Code\":0,\"Msg\":\"suc\"}", Tid)
}

func AllGrp(Tid uint32, req string) string {
	defer profiler("AllGrp", time.Now().UnixNano())

	var reqStruct AllGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(req)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	resp, err := GetAllGrp(Uname)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	resp.Tid = Tid
	respBody, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(resp)\"}", Tid)
	}

	return string(respBody)
}

func CreateGrp(Tid uint32, req string) string {
	defer profiler("CreateGrp", time.Now().UnixNano())

	var reqStruct CreateGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	//check Grp NUM
	if uint16(len(AllGrp.Grps)) >= g_conf.UserMaxGrp {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"More than the largest number of groups\"}", Tid)
	}

	//check GrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.GrpName, V.GrpName) {
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"already existed\"}", Tid)
		}
	}

	storeData := &StoreFormat{Order: len(AllGrp.Grps), GrpID: uint64(time.Now().UnixNano() / 1000000), GrpName: reqStruct.GrpName, GrpVer: 0, Items: []string{}}
	err = SetGrpData(Uname, storeData.GrpID, storeData)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	resp := &CreateGrp_resp{Tid: Tid, Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail\"}", Tid)
	}

	return string(respBody)
}

func DeleteGrp(Tid uint32, req string) string {
	defer profiler("DeleteGrp", time.Now().UnixNano())

	var reqStruct DeleteGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	err = DelGrp(Uname, reqStruct.GrpIDs)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	return fmt.Sprintf("{\"Tid\":%d,\"Code\":0,\"Msg\":\"suc\"}", Tid)
}

func RenameGrp(Tid uint32, req string) string {
	defer profiler("RenameGrp", time.Now().UnixNano())

	var reqStruct RenameGrp_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	AllGrp, err := GetAllGrp(Uname)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	//check NewGrpName
	for _, V := range AllGrp.Grps {
		if strings.EqualFold(reqStruct.NewGrpName, V.GrpName) {
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"already existed\"}", Tid)
		}
	}

	storeData.GrpName = reqStruct.NewGrpName
	//storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	resp := &RenameGrp_resp{Tid: Tid, Code: 0, Msg: "suc", GrpID: reqStruct.GrpID, GrpVer: storeData.GrpVer}
	respBody, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail\"}", Tid)
	}

	return string(respBody)
}

func ChangeGrpOrder(Tid uint32, req string) string {
	defer profiler("ChangeOrderGrp", time.Now().UnixNano())

	var reqStruct ChangeGrpOrder_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	for i, GrpID := range reqStruct.GrpOrder {

		storeData, err := GetGrpData(Uname, GrpID)
		if err != nil {
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
		}

		storeData.Order = i
		err = SetGrpData(Uname, GrpID, storeData)
		if err != nil {
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
		}
	}

	return fmt.Sprintf("{\"Tid\":%d,\"Code\":0,\"Msg\":\"suc\"}", Tid)
}

func Upload(Tid uint32, req string) string {
	defer profiler("Upload", time.Now().UnixNano())

	var reqStruct Upload_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	bReturnItems := false
	if reqStruct.GrpVer != storeData.GrpVer {
		bReturnItems = true
	}

	switch reqStruct.Action {
	case 0:
		if uint16(len(storeData.Items))+uint16(len(reqStruct.Items)) > g_conf.GrpMaxItem {
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"More than the largest number of item\"}", Tid)
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
			return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"More than the largest number of item\"}", Tid)
		}
		storeData.Items = reqStruct.Items
	}
	storeData.GrpVer = uint64(time.Now().UnixNano() / 1000000)

	err = SetGrpData(Uname, reqStruct.GrpID, storeData)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	var respBody []byte
	if bReturnItems {
		resp := &Download_resp{Tid: Tid, Code: 2, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
		resp.Items = storeData.Items
		respBody, err = json.Marshal(resp)
	} else {
		resp := &Upload_resp{Tid: Tid, Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
		respBody, err = json.Marshal(resp)
	}

	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail\"}", Tid)
	}

	return string(respBody)
}

func Download(Tid uint32, req string) string {
	defer profiler("Download", time.Now().UnixNano())
	var reqStruct Download_req
	err := json.Unmarshal([]byte(req), &reqStruct)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(data parse)\"}", Tid)
	}

	Uname, err := CheckSid(reqStruct.Sid)
	if err != nil || len(Uname) == 0 {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(Sid)\"}", Tid)
	}

	storeData, err := GetGrpData(Uname, reqStruct.GrpID)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail(db)\"}", Tid)
	}

	if reqStruct.GrpVer == storeData.GrpVer {
		respStr := fmt.Sprintf("{\"Tid\":%d,\"Code\":1,\"Msg\":\"isLatest\",\"GrpID\":%d, \"GrpVer\":%d}",
			Tid, storeData.GrpID, storeData.GrpVer)
		return respStr
	}

	resp := &Download_resp{Tid: Tid, Code: 0, Msg: "suc", GrpID: storeData.GrpID, GrpVer: storeData.GrpVer}
	resp.Items = storeData.Items
	respBody, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("{\"Tid\":%d,\"Code\":-1,\"Msg\":\"fail\"}", Tid)
	}
	return string(respBody)
}
