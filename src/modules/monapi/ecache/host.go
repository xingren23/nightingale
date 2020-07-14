package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

// ip -> 主机信息
type HostCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.CmdbHost
}

var HostCache *HostCacheMap

func NewHostCache() *HostCacheMap {
	return &HostCacheMap{
		Data: make(map[string]*dataobj.CmdbHost),
	}
}

func (this *HostCacheMap) Get(key string) (*dataobj.CmdbHost, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}

func (this *HostCacheMap) SetAll(vals []*dataobj.CmdbHost) {
	if vals == nil || len(vals) == 0 {
		return
	}
	m := make(map[string]*dataobj.CmdbHost)
	for _, host := range vals {
		if host != nil {
			m[host.Ip] = host
		}
	}
	this.Lock()
	defer this.Unlock()
	this.Data = m
}

func (this *HostCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
