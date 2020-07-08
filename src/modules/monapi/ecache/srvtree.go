package ecache

import (
	"sync"
)

// 服务树节点id -> 节点串
type SrvTreeCacheMap struct {
	sync.RWMutex
	Data map[int64]string
}

var SrvTreeCache *SrvTreeCacheMap

func NewSrvTreeCache() *SrvTreeCacheMap {
	return &SrvTreeCacheMap{
		Data: make(map[int64]string),
	}
}

func (this *SrvTreeCacheMap) Get(key int64) (string, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[key]
	return value, exists
}

func (this *SrvTreeCacheMap) SetAll(data map[int64]string) {
	this.Lock()
	defer this.Unlock()
	this.Data = data
}