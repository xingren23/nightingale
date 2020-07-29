package cache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
)

// endpoint cache : ident -> endpoint
type EndpointCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.Endpoint
}

func NewEndpointCache() *EndpointCacheMap {
	return &EndpointCacheMap{
		Data: make(map[string]*dataobj.Endpoint),
	}
}

func (this *EndpointCacheMap) Get(key string) (*dataobj.Endpoint, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[key]
	return value, exists
}

func (this *EndpointCacheMap) SetAll(m map[string]*dataobj.Endpoint) {
	this.Lock()
	defer this.Unlock()
	this.Data = m
}
