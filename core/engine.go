package core

//引擎控制

import (
	"strconv"
	"sync"

	"github.com/hellobchain/nginxgo/common/constant"
	"github.com/hellobchain/nginxgo/common/log"
	"github.com/hellobchain/wswlog/wlogging"
)

// 日志
var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

// 引擎
type Engine struct {
	service           []service
	upstream          map[string]*upstream
	servicesPoll      map[string]*service //现有的服务池
	resetServicesPoll map[string]*service //重启后的服务池
	state             int                 //引擎现在的状态
	wg                sync.WaitGroup
}

func createEngine() *Engine {
	engine := Engine{}
	engine.resetServicesPoll = make(map[string]*service)
	engine.servicesPoll = make(map[string]*service)
	return &engine
}

func (engine *Engine) writeEngine(cfg config) {
	engine.service = cfg.Service
	engine.upstream = cfg.Upstream

	//处理后端服务器池，建构哈希环
	for _, v := range engine.upstream {
		v.mu.Lock()
		v.hashRing = &hashRing{}
		v.hashRing.nodes = make(map[int]string)
		v.addNode()
		v.mu.Unlock()
	}

	//处理服务节点
	for i := range engine.service {
		service := &engine.service[i]
		for _, location := range service.Location {
			//计算location哈希值，用于一致性比对
			location.hashValue = hash([]byte(strconv.Itoa(location.LocationType) + location.Root + location.FileRoot + location.Upstream))
			service.hashValue += uint64(location.hashValue)
		}
		if engine.state == constant.ENGINE_RESET { //reset信息写入reset map
			engine.resetServicesPoll[service.Port] = service
		}
	}
}

func (engine *Engine) resetEngine() {
	engine.state = constant.ENGINE_RESET
	readConfig(engine)
	for key, value := range engine.resetServicesPoll {
		//首先确定不存在的，启动服务
		src, ok := engine.servicesPoll[key]
		if !ok {
			engine.wg.Add(1)
			go value.listen(&engine.servicesPoll, &engine.upstream, &engine.wg)
			continue
		}
		//如果已经存在，则确认哈希value是否是一致的
		if value.hashValue != src.hashValue {
			src.mu.Lock()
			//如果不等于，则停掉原来的服务，再根据新的重启
			delete(engine.servicesPoll, key)
			err := src.httpService.Close()
			if err != nil {
				logger.Error("关闭服务错误：", err)
			}
			engine.wg.Add(1)
			go value.listen(&engine.servicesPoll, &engine.upstream, &engine.wg)
			src.mu.Unlock()
		}
	}
	//确认已经关掉的服务
	for key, value := range engine.servicesPoll {
		_, ok := engine.resetServicesPoll[key]
		if !ok {
			delete(engine.servicesPoll, key)
			err := value.httpService.Close()
			if err != nil {
				logger.Error("关闭服务错误：", err)
			}
		}
	}
	engine.resetServicesPoll = make(map[string]*service) //释放内存
}

func (engine *Engine) stopEngine() {
	for _, value := range engine.servicesPoll {
		value.httpService.Close()
	}
	logger.Info("程序退出")
}
