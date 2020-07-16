package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type NetworkCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.Network
}

var NetworkCache *NetworkCacheMap

func NewNetworkCache() *NetworkCacheMap {
	return &NetworkCacheMap{
		Data: map[string]*dataobj.Network{},
	}
}

func (this *NetworkCacheMap) SetAll(vals map[string]*dataobj.Network) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *NetworkCacheMap) GetByIp(ip string) (*dataobj.Network, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *NetworkCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
