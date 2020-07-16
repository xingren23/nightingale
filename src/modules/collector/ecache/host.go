package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type HostCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.CmdbHost
}

var HostCache *HostCacheMap

func NewHostCache() *HostCacheMap {
	return &HostCacheMap{
		Data: map[string]*dataobj.CmdbHost{},
	}
}

func (this *HostCacheMap) GetByIp(ip string) (*dataobj.CmdbHost, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *HostCacheMap) SetAll(vals map[string]*dataobj.CmdbHost) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *HostCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
