package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

type HostCacheMap struct {
	sync.RWMutex
	Data map[string]*meicai.CmdbHost
}

var HostCache *HostCacheMap

func NewHostCache() *HostCacheMap {
	return &HostCacheMap{
		Data: map[string]*meicai.CmdbHost{},
	}
}

func (this *HostCacheMap) GetByIp(ip string) (*meicai.CmdbHost, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *HostCacheMap) SetAll(vals map[string]*meicai.CmdbHost) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *HostCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
