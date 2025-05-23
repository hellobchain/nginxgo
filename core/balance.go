package core

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 负载均衡器
// 哈希环建构
type hashRing struct {
	ring  []int          //哈希环
	nodes map[int]string //节点哈希映射到节点名称
}

// 新增哈希节点，replicas为每个真实节点对应的虚拟节点数
func (upstream *upstream) addNode() {
	for _, node := range upstream.Addr {
		for i := 0; i < upstream.Replicas; i++ {
			hashValue := int(hash([]byte(strconv.Itoa(i) + node)))
			upstream.hashRing.ring = append(upstream.hashRing.ring, hashValue)
			upstream.hashRing.nodes[hashValue] = node
		}
	}
	sort.Ints(upstream.hashRing.ring)
}

// 均衡器。利用客户端ip计算客户端的哈希值，并且获取顺时针的节点。通过二分查找进行。
func (hashRing hashRing) balancer(ip string) string {
	hash := int(hash([]byte(ip)))
	idx := sort.Search(len(hashRing.ring), func(i int) bool {
		return hashRing.ring[i] >= hash
	})
	if idx == len(hashRing.ring) {
		idx = 0
	}
	return hashRing.nodes[hashRing.ring[idx]]
}

// 计算crc
func hash(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// 从后端服务器池中删除字段并重构哈希环
func (upstream *upstream) del(ip string) {
	for i, v := range upstream.Addr {
		//删除切片中的指定元素
		if ip == v {
			upstream.Addr = append(upstream.Addr[:i], upstream.Addr[i+1:]...)
			break
		}
	}
	upstream.mu.Lock()
	upstream.hashRing = &hashRing{}
	upstream.hashRing.nodes = make(map[int]string)
	upstream.addNode()
	upstream.mu.Unlock()
}
