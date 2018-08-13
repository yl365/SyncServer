# SyncServer
一个支持http/Websocket协议的多列表同步服务器

## 特性
1.　支持多协议； 
  >基于http的RESTful协议；    
  >基于Websocket的JOSN协议；      
  
2.　支持多组同步；支持版本比较下载，减少流量；

3.　权限/存储基于redis，与业务服务器分离，方便扩展；

4.　每个消息有Tid,支持异步模式；


## 获取
```
go get -u github.com/yl365/SyncServer
```

依赖的第三方库:
```
go get -u github.com/gobwas/ws
go get -u github.com/gobwas/ws/wsutil
go get -u github.com/gomodule/redigo/redis
```

可考虑优化:
  >JSON解析部分,采用系统默认的库,可考虑换成第三方更高性能的库；

  >为了复用数据解析部分，http协议数据部分，直接采用了json格式;


## 使用

1、 修改配置文件config.toml

```
# 配置文件, utf8编码, 格式遵循TOML配置规范
# https://github.com/toml-lang/toml#user-content-spec

# 日志级别 debug, info, error
LogLevel = 'debug'
# 如果开启,将在程序的第一分钟做Profile,方便分析程序
EnableProfile = false

listen = ':9995'
# 同一IP请求频率限制: 5分钟内最大请求次数, 超过将限制请求5分钟; 0不限制
ReqFreqLimit = 0

# 数据存储redis配置
RedisServer = '127.0.0.1:6379'
RedisPasswd = ''
RedisDB = 0
# RedisKey格式: HGETALL D:Uname 
RedisKey = 'D:%s'

# 用户名密码验证的redis, 可以和上面是同一个
UserRedisServer = '127.0.0.1:6379'
UserRedisPasswd = ''
UserRedisDB = 0
# UserRedisKey格式: GET U:Uname 
UserRedisKey = 'U:%s'

# 业务相关配置: 每个用户最多多少组, 每个组最多多少条目
UserMaxGrp = 10
GrpMaxItem = 100

# 会话ID超时时间, 单位秒, 当超过这个时间没有请求会被清理
SidTimeOut = 3600
```

2、 在配置目录运行: nohup ./SyncServer &



## 存储格式

	密码KV:   U:uname passwd
	数据HASH: D:uname GrpID  {"Order":n,"GrpID":nnn,"GrpName":"xxx","GrpVer":nnn,"Items":["SH600001","SH600002"]}

## 协议细节


请求:

	http://v1.domain.com:port/api?req=URLEncode(json) (支持GET/POST)

	ws://v1.domain.com:port/ws
	
响应:

	{json}

	
登入:

	请求: {"Tid":n,"Type":"Login","Uname":"xxx","Passwd":"xxx"}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail","Sid":nnn}

	
登出:

	请求: {"Tid":n,"Type":"Logout","Sid":nnn}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail"}

	
获取全部分组:

	请求: {"Tid":n,"Type":"AllGrp","Sid":nnn}

	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail","Groups":[[组ID,组名,版本],[组ID,组名,版本]...]}

	
创建新分组:

	请求: {"Tid":n,"Type":"CreateGrp","Sid":nnn,"GrpName":"xxx"}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn}

	
删除分组:

	请求: {"Tid":n,"Type":"DeleteGrp","Sid":nnn,"GrpIDs":[nnn,nnn...]}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail/..."}

	
修改组名:

	请求: {"Tid":n,"Type":"RenameGrp","Sid":nnn,"GrpID":nnn,"NewGrpName":"xxx"}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail/...","GrpID":nnn,"GrpVer":nnn}

	
调整组顺序:

	请求: {"Tid":n,"Type":"ChangeGrpOrder","Sid":nnn,"GrpOrder":[nnn,nnn...]}
	
	返回: {"Tid":n,"Code":0/-1,"Msg":"suc/fail/..."}

	
上传条目(添加/删除/覆盖):

	请求: {"Tid":n,"Type":"Upload","Sid":nnn,"Action":0/1/2,"GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
	
	返回: {"Tid":n,"code":0,"msg":"suc","GrpID":nnn,"GrpVer":nnn}
	
		  {"Tid":n,"Code":0,"Msg":"suc","GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
		
		  {"Tid":n,"Code":-1,"Msg":"fail(...)"}
		  
下载条目:

	请求: {"Tid":n,"Type":"Download","Sid":nnn,"GrpID":nnn,"GrpVer":nnn}
	
	返回: {"Tid":n,"Code":0,"Msg":"isLatest","GrpID":nnn,"GrpVer":nnn}
	
		　{"Tid":n,"Code":0,"Msg":"suc","GrpID":nnn,"GrpVer":nnn,"Items":["SH600001","SH600002"...]}
		
		　{"Tid":n,"Code":-1,"Msg":"fail(...)"}


欢迎试用并提出意见建议。如果发现bug，请Issues，谢谢！

如果你觉得这个项目有意义，或者对你有帮助，或者仅仅是为了给我一点鼓励，请不要吝惜，给我一个*star*，谢谢！≡ω≡

