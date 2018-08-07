package main

/*
	请求(GET/POST):
		http://ip:port/api/v1/AllGrp?req=URLEncode(json)
	响应:{json}

	存储格式:
		密码KV:   U:uname passwd
		数据HASH: D:uname GrpID  {"Order":n,"GrpID":nnn,"GrpName":"xxx","GrpVer":nnn,"Items":["SH600001","SH600002"]}
*/

type StoreFormat struct {
	Order   int      `json:"Order"`
	GrpID   uint64   `json:"GrpID"`
	GrpVer  uint64   `json:"GrpVer"`
	GrpName string   `json:"GrpName"`
	Items   []string `json:"Items"`
}

/*
	协议格式: {"Type":"xxx",...}
*/
type DataPackage struct {
	Type string `json:"Type"`
}

/*
	登入:
		请求: {"Type":"xxx","Uname":"xxx","Passwd":"xxx"}
		返回: {"Code":0,"Msg":"suc/fail","Sid":nnn}
*/
type Login_req struct {
	Type   string `json:"Type"`
	Uname  string `json:"Uname"`
	Passwd string `json:"Passwd"`
}

type Login_resp struct {
	Code int    `json:"Code"`
	Msg  string `json:"Msg"`
	Sid  uint64 `json:"Sid"`
}

/*
	登出:
		请求: {"Sid":nnn}
		返回: {"Code":0,"Msg":"suc/fail"}
*/
type Logout_req struct {
	Sid uint64 `json:"Sid"`
}

type Logout_resp struct {
	Code int    `json:"Code"`
	Msg  string `json:"Msg"`
}

/*
	获取全部分组:
		请求: {"Sid":nnn}
		返回: {"Code":0,"Msg":"suc/fail","Groups":[[组ID,组名,版本],[组ID,组名,版本]...]}
*/
type AllGrp_req struct {
	Sid uint64 `json:"Sid"`
}
type GrpInfo struct {
	Order   int    `json:"Order"`
	GrpID   uint64 `json:"GrpID"`
	GrpVer  uint64 `json:"GrpVer"`
	GrpName string `json:"GrpName"`
}
type GrpInfoSlice []GrpInfo
type AllGrp_resp struct {
	Code int          `json:"Code"`
	Msg  string       `json:"Msg"`
	Grps GrpInfoSlice `json:"Groups"`
}

func (c GrpInfoSlice) Len() int {
	return len(c)
}
func (c GrpInfoSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c GrpInfoSlice) Less(i, j int) bool {
	return c[i].Order < c[j].Order
}

/*
	创建新分组:
		请求: {"Sid":nnn,"GrpName":"xxx"}
		返回: {"Code":0,"Msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn}
*/

type CreateGrp_req struct {
	Sid     uint64 `json:"Sid"`
	GrpName string `json:"GrpName"`
}

type CreateGrp_resp struct {
	Code   int    `json:"Code"`
	Msg    string `json:"Msg"`
	GrpID  uint64 `json:"GrpID"`
	GrpVer uint64 `json:"GrpVer"`
}

/*
	删除分组:
		请求: {"Sid":nnn,"GrpIDs":[nnn,nnn...]}
		返回: {"Code":0,"Msg":"suc/fail/..."}
*/

type DeleteGrp_req struct {
	Sid    uint64   `json:"Sid"`
	GrpIDs []uint64 `json:"GrpIDs"`
}

type DeleteGrp_resp struct {
	Code int    `json:"Code"`
	Msg  string `json:"Msg"`
}

/*
	修改组名:
		请求: {"Sid":nnn,"GrpID":nnn,"NewGrpName":"xxx"}
		返回: {"Code":0,"Msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn}
*/
type RenameGrp_req struct {
	Sid        uint64 `json:"Sid"`
	GrpID      uint64 `json:"GrpID"`
	NewGrpName string `json:"NewGrpName"`
}

type RenameGrp_resp struct {
	Code   int    `json:"Code"`
	Msg    string `json:"Msg"`
	GrpID  uint64 `json:"GrpID"`
	GrpVer uint64 `json:"GrpVer"`
}

/*
	调整组顺序[预留, 暂不需要]:
		请求: {"Sid":nnn,"GrpOrder":[nnn,nnn...]}
		返回: {"Code":0,"Msg":"suc/fail/..."}
*/

type ChangeGrpOrder_req struct {
	Sid      uint64   `json:"Sid"`
	GrpOrder []uint64 `json:"GrpOrder"`
}

type ChangeGrpOrder_resp struct {
	Code int    `json:"Code"`
	Msg  string `json:"Msg"`
}

/*
	上传自选股(添加/删除/覆盖):
		请求: {"Sid":nnn,"Action":0/1/2,"GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
		返回: {"code":0,"msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn}
			  {"Code":0,"Msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
*/

type Upload_req struct {
	Sid    uint64   `json:"Sid"`
	Action uint8    `json:"Action"`
	GrpID  uint64   `json:"GrpID"`
	GrpVer uint64   `json:"GrpVer"`
	Items  []string `json:"Items"`
}

type Upload_resp struct {
	Code   int    `json:"Code"`
	Msg    string `json:"Msg"`
	GrpID  uint64 `json:"GrpID"`
	GrpVer uint64 `json:"GrpVer"`
}

/*
	下载自选股:
		请求: {"Sid":nnn,"GrpID":nnn,"GrpVer":nnn}
		返回: {"Code":0,"Msg":"suc/fail/isLatest","GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
*/

type Download_req struct {
	Sid    uint64 `json:"Sid"`
	GrpID  uint64 `json:"GrpID"`
	GrpVer uint64 `json:"GrpVer"`
}

type Download_resp struct {
	Code   int      `json:"Code"`
	Msg    string   `json:"Msg"`
	GrpID  uint64   `json:"GrpID"`
	GrpVer uint64   `json:"GrpVer"`
	Items  []string `json:"Items"`
}
