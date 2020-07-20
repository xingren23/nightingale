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
	defer this.RUnlock()
	value, exists := this.Data[key]
	return value, exists
}

func (this *EndpointCacheMap) SetAll(m map[string]*model.Endpoint) {
	this.Lock()
	defer this.Unlock()
	this.Data = m
}