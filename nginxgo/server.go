package nginxgo

// Init 初始化服务，需要提供哈希环上每个真实节点对应的虚拟节点个数
func Init() *Engine {
	e := createEngine()
	readConfig(e)
	return e
}

// Start 启动服务
func (e *Engine) Start() {
	e.startListen()
	e.wg.Wait()
}

// Reset 重启动服务，不中断服务。
func (e *Engine) Reset() {
	e.resetEngine()
	e.wg.Wait()
}

func (e *Engine) Stop() {
	e.stopEngine()
}
