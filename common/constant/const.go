package constant

const (
	CMD_START = "start"
	CMD_RESET = "reset"
	CMD_STOP  = "stop"
	CMD_HELP  = "help"
)

// location type描述
const (
	LOCATION_LOADBALANCING = 1
	LOCATION_FILESERVICE   = 2
)

// 引擎状态描述
const (
	ENGINE_START = 1
	ENGINE_RUN   = 2
	ENGINE_RESET = 3 //实际有用的好像就这个
)

// block
const (
	BLOCK_SERVER      = "[server]"
	BLOCK_SERVER_PORT = "port"

	BLOCK_UPSTREAM          = "[upstream]"
	BLOCK_UPSTREAM_NAME     = "name"
	BLOCK_UPSTREAM_SCHEMA   = "schema"
	BLOCK_UPSTREAM_REPLICAS = "replicas"

	BLOCK_LOCATION           = "[location]"
	BLOCK_LOCATION_TYPE      = "type"
	BLOCK_LOCATION_ROOT      = "root"
	BLOCK_LOCATION_UPSTREAM  = "upstream"
	BLOCK_LOCATION_FILE_ROOT = "file_root"

	BLOCK_PROXY_SET_HEADER       = "[proxy_set_header]"
	BLOCK_PROXY_SET_HEADER_KEY   = "key"
	BLOCK_PROXY_SET_HEADER_VALUE = "value"
	BLOCK_END                    = "[end]"
)

// 日志切割默认配置
const (
	DEFAULT_MAX_AGE       = 7   // 日志最长保存时间，单位：天
	DEFAULT_ROTATION_TIME = 24  // 日志滚动间隔，单位：小时
	DEFAULT_ROTATION_SIZE = 100 // 默认的日志滚动大小，单位：MB
)
