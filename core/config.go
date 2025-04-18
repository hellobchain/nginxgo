package core

//读取配置文件

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hellobchain/nginxgo/common/constant"
)

// 配置文件结构
type config struct {
	Service  []service            `json:"service"`
	Upstream map[string]*upstream `json:"upstream"` //一个后端服务器池名对应多个后端服务器
}

// upstream结构
type upstream struct {
	Addr           []string          `json:"addr"` //服务器地址
	hashRing       *hashRing         //upstream对应的哈希环
	Replicas       int               `json:"replicas"` //每个虚拟节点对应的真实节点数量
	Scheme         string            `json:"scheme"`   //协议
	failCount      map[string]int    //记录每个后端服务器的失败连接次数，超过三次就把这个服务器从这个池子里扬了
	mu             sync.Mutex        //一把锁，用于热重启
	ProxySetHeader []*proxySetHeader `json:"proxy_set_header"` // 代理请求头
}

// service结构
type service struct {
	Port        string       `json:"port"` //定义监听的代理服务器端口号。一个端口号绑定一个service。
	httpService *http.Server //定义一个http服务，用于启动代理服务器
	Location    []*location  `json:"location"` //location结构
	hashValue   uint64       //location哈希值的和。
	mu          sync.Mutex   // 一把锁，用于热重启
}

// location结构
type location struct {
	LocationType int    `json:"type"`      // location类型，分为两种，一种是文件服务，一种是负载均衡服务
	Root         string `json:"root"`      //根路径，会附加在service结构的根路径上
	Upstream     string `json:"upstream"`  //使用的后端服务器池名
	FileRoot     string `json:"file_root"` //fileRoot，文件路径，和root是两个东西了
	hashValue    uint32 //location哈希值。用于验证location是否发生变化。
}

// proxySetHeader结构体
type proxySetHeader struct {
	HeaderName  string `json:"key"`   // header名称
	HeaderValue string `json:"value"` // header值
}

func readConfigFromFile(fileName string) config {
	var cfg config
	cfg.Upstream = make(map[string]*upstream)
	file, err := os.Open(fileName)
	if err != nil {
		logger.Fatalf("open config file failed: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	//定义当前的type
	const (
		serviceType  = 1
		upstreamType = 2
		locationType = 3
		proxyType    = 4
		endType      = 0
	)
	var nowType = 0
	var serviceStruct service
	var locationStruct location
	var proxyStruct proxySetHeader
	var upstreamName string
	for scanner.Scan() {
		line := scanner.Text()
		if isSkip(line) {
			continue
		}
		//检测block
		switch line {
		case constant.BLOCK_SERVER:
			nowType = serviceType
			continue
		case constant.BLOCK_UPSTREAM:
			nowType = upstreamType
			continue
		case constant.BLOCK_LOCATION:
			nowType = locationType
			continue
		case constant.BLOCK_PROXY_SET_HEADER:
			nowType = proxyType
			continue
		case constant.BLOCK_END:
			switch nowType {
			case serviceType:
				cfg.Service = append(cfg.Service, serviceStruct)
				// 重置结构体
				serviceStruct = service{}
				nowType = endType
			case locationType:
				//复制一份
				newLocation := locationStruct
				serviceStruct.Location = append(serviceStruct.Location, &newLocation)
				locationStruct = location{}
				nowType = serviceType
			case upstreamType:
				upstreamName = ""
				nowType = endType
			case proxyType:
				//复制一份
				newProxySetHeader := proxyStruct
				cfg.Upstream[upstreamName].ProxySetHeader = append(cfg.Upstream[upstreamName].ProxySetHeader, &newProxySetHeader)
				proxyStruct = proxySetHeader{}
				nowType = upstreamType
			}
			continue
		}
		//检查目前字段所处区块
		switch nowType {
		//处理service区块
		case serviceType:
			s := strings.Split(line, "=")
			switch s[0] {
			case constant.BLOCK_SERVER_PORT:
				serviceStruct.Port = s[1]
			}
		case upstreamType:
			s := strings.Split(line, "=")
			switch s[0] {
			case constant.BLOCK_UPSTREAM_NAME:
				upstreamName = s[1]
				cfg.Upstream[upstreamName] = &upstream{}
				cfg.Upstream[upstreamName].failCount = make(map[string]int)
			case constant.BLOCK_UPSTREAM_REPLICAS:
				replicas, err := strconv.Atoi(s[1])
				if err != nil {
					logger.Fatalf("replicas 字段设置错误：%v", err)
				}
				cfg.Upstream[upstreamName].Replicas = replicas
			case constant.BLOCK_UPSTREAM_SCHEMA:
				cfg.Upstream[upstreamName].Scheme = s[1]
			default:
				cfg.Upstream[upstreamName].Addr = append(cfg.Upstream[upstreamName].Addr, s[0])
			}
		case locationType:
			s := strings.Split(line, "=")
			switch s[0] {
			case constant.BLOCK_LOCATION_TYPE:
				typeNum, err := strconv.Atoi(s[1])
				if err != nil {
					logger.Fatalf("location type字段设置错误:%v", err)
				}
				locationStruct.LocationType = typeNum
			case constant.BLOCK_LOCATION_ROOT:
				locationStruct.Root = s[1]
			case constant.BLOCK_LOCATION_UPSTREAM:
				locationStruct.Upstream = s[1]
			case constant.BLOCK_LOCATION_FILE_ROOT:
				locationStruct.FileRoot = s[1]
			}
		case proxyType:
			s := strings.Split(line, "=")
			switch s[0] {
			case constant.BLOCK_PROXY_SET_HEADER_KEY:
				proxyStruct.HeaderName = s[1]
			case constant.BLOCK_PROXY_SET_HEADER_VALUE:
				proxyStruct.HeaderValue = s[1]
			}
		}
	}
	printJsonCfg(cfg)
	return cfg
}

func printJsonCfg(cfg config) {
	ret, _ := json.MarshalIndent(cfg, "", " ")
	logger.Infof("nginxgo config: %s", string(ret))
}

// 读取配置文件
func readConfig(engine *Engine) {
	engine.writeEngine(readConfigFromFile("./configs/config.cfg"))
}

// 跳过检测。
func isSkip(line string) bool {
	if line == "" || line == " " || line == "\n" || line == "\r" || line == "\t" || line[0] == '#' {
		return true
	}
	return false
}
