package ecache

import (
	"sync"
)

// 服务树节点id -> 节点串
type SrvTreeCacheMap struct {
	sync.RWMutex
	Data map[int64]string
}

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

func (this *SrvTreeCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *SrvTreeCacheMap) GetAll() map[int64]string {
	this.Lock()
	defer this.Unlock()

	m := make(map[int64]string, len(this.Data))
	for k, v := range this.Data {
		m[k] = v
	}
	return m
}
