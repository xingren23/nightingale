package ecache

import "sync"

// 服务树节点串-> endpoint列表
type StraEndpointCacheMap struct {
	sync.RWMutex
	Data map[string][]string
}

var StraEndpointCache *StraEndpointCacheMap

func NewStraEndpointCache() *StraEndpointCacheMap {
	return &StraEndpointCacheMap{
		Data: make(map[string][]string),
	}
}

func (this *StraEndpointCacheMap) Get(key string) ([]string, bool) {
	this.RLock()
	value, exists := this.Data[key]
	this.RUnlock()
	return value, exists
}
