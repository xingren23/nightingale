package ecache

import (
	"github.com/didi/nightingale/src/model"
	"sync"
)

// endpoint -> endpoint信息
type EndpointCacheMap struct {
	sync.RWMutex
	Data map[string]*model.Endpoint
}

var EndpointCache *EndpointCacheMap

func NewEndpointCache() *EndpointCacheMap {
	return &EndpointCacheMap{
		Data: make(map[string]*model.Endpoint),
	}
}

func (this *EndpointCacheMap) Get(key string) (*model.Endpoint, bool) {
	this.RLock()
	value, exists := this.Data[key]
	this.RUnlock()
	return value, exists
}
