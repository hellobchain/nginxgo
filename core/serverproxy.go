package core

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/hellobchain/nginxgo/common/constant"
	"github.com/hellobchain/nginxgo/pkg/uuid"
)

//实现反向代理

// 启动监听服务
func (engine *Engine) startListen() {
	for i := range engine.service {
		service := &engine.service[i]
		engine.wg.Add(1)
		go service.listen(&engine.servicesPoll, &engine.upstream, &engine.wg)
	}
}

// 对每个service进行监听
func (service *service) listen(servicesPoll *map[string]*service, upstreamMap *map[string]*upstream, wg *sync.WaitGroup) {
	mux := http.NewServeMux()
	for i := range service.Location {
		location := &service.Location[i]
		if (*location).Root == "" {
			(*location).Root = "/"
		}
		switch (*location).LocationType {
		case constant.LOCATION_LOADBALANCING:
			mux.HandleFunc((*location).Root, func(writer http.ResponseWriter, request *http.Request) {
				(*location).forward(writer, request, &service.mu, upstreamMap)
			})
		case constant.LOCATION_FILESERVICE:
			mux.HandleFunc((*location).Root, func(writer http.ResponseWriter, request *http.Request) {
				(*location).getFile(writer, request, &service.mu)
			})
		}
	}

	src := &http.Server{
		Addr:    ":" + service.Port, //还是和端口绑定了，令人感叹
		Handler: mux,
	}
	service.httpService = src
	(*servicesPoll)[service.Port] = service

	wg.Done()

	err := src.ListenAndServe()
	if err != nil {
		if err.Error() == "http: Server closed" {
			return
		}
		logger.Error("监听", service.Port, "错误，错误信息：", err)
	}
}

// 反向代理，将信息转发给后端服务器
func (location *location) forward(w http.ResponseWriter, r *http.Request, mu *sync.Mutex, upstreamMap *map[string]*upstream) {
	logRequest(r)
	//询问是否正在热重启。如果是则返回503，服务器维护状态码。
	isNotReSet := mu.TryLock()
	if !isNotReSet {
		http.Error(w, "服务重启中，请重试", http.StatusServiceUnavailable)
		return
	} else {
		mu.Unlock()
	}

	//	获取hash环
	upstream := (*upstreamMap)[location.Upstream]
	isNotReSet = upstream.mu.TryLock()
	if !isNotReSet {
		http.Error(w, "服务重启中，请重试", http.StatusServiceUnavailable)
		return
	} else {
		upstream.mu.Unlock()
	}
	hashRing := upstream.hashRing

	// 获取客户端ip
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		logger.Error("获取ip错误:", err)
		http.Error(w, "获取ip错误", http.StatusInternalServerError)
		return
	}

	// 获取后端服务器
	serviceIP := hashRing.balancer(ip)
	remote, err := url.Parse(upstream.Scheme + "://" + serviceIP)
	if err != nil {
		logger.Error("解析目标服务器地址失败:", err)
		http.Error(w, "解析目标服务器地址失败", http.StatusInternalServerError)
		return
	}

	// 创建反向代理。
	proxy := httputil.NewSingleHostReverseProxy(remote)
	// 修改响应头
	proxy.ModifyResponse = func(resp *http.Response) error {
		for _, header := range upstream.ProxySetHeader {
			w.Header().Add(header.HeaderName, header.HeaderValue)
		}
		return nil
	}
	proxy.ServeHTTP(w, r)
	if len(w.Header()) == 0 {
		upstream.failCount[serviceIP] += 1
		count := upstream.failCount[serviceIP]
		if count == 3 {
			logger.Error("后端服务器", serviceIP, "已失效")
			upstream.del(serviceIP)
		}
	}
}

func logRequest(r *http.Request) {
	requestId := uuid.GetUUIDInt()
	// 开始时间
	start := time.Now()
	// path
	path := r.URL.Path
	// ip
	clientIP := r.RemoteAddr
	// 方法
	method := r.Method
	// 结束时间
	end := time.Now()
	// 执行时间
	latency := end.Sub(start)
	logger.Infof("| %10d | %13v | %15s | %s  %s |",
		requestId,
		latency,
		clientIP,
		method,
		path,
	)

}
